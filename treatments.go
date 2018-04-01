package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MJKWoolnough/errors"
	"github.com/MJKWoolnough/memio"
)

var Treatments treatments

type Treatment struct {
	ID             uint
	Name           string
	Category       uint
	Price          uint
	Duration       time.Duration
	Order          uint
	DescriptionSrc string
	Description    *PageBytes
}

type Category struct {
	ID        uint
	Name      string
	Order     uint
	AdminOnly bool
}

type treatments struct {
	addTreatment, updateTreatment, removeTreatment *sql.Stmt
	addCategory, updateCategory, removeCategory    *sql.Stmt

	mu           sync.RWMutex
	treatments   map[uint]Treatment
	categories   map[uint]Category
	categoryT    *template.Template
	categoryPage *PageBytes
}

func (t *treatments) Init(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS [Treatment]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Category] INTEGER NOT NULL, [Price] INTEGER NOT NULL, [Duration] INTEGER NOT NULL, [Description] TEXT NOT NULL, [Order] INTEGER NOT NULL);")
	if err != nil {
		return errors.WithContext("error creating Treatment table: ", err)
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS [Category]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Order] INTEGER DEFAULT 0 NOT NULL, [AdminOnly] BOOLEAN DEFAULT 0 NOT NULL);")
	if err != nil {
		return errors.WithContext("error creating Category table: ", err)
	}

	for _, stmt := range [...]struct {
		Stmt  **sql.Stmt
		Query string
	}{
		{&t.addTreatment, "INSERT INTO [Treatment]([Name], [Category], [Price], [Duration], [Description], [Order]) VALUES (?, ?, ?, ?, ?, ?);"},
		{&t.updateTreatment, "UPDATE [Treatment] SET [Name] = ?, [Category] = ?, [Price] = ?, [Duration] = ?, [Description] = ?, [Order] = ? WHERE [ID] = ?;"},
		{&t.removeTreatment, "DELETE FROM [Treatment] WHERE [ID] = ?;"},
		{&t.addCategory, "INSERT INTO [Category]([Name], [Order], [AdminOnly]) VALUES (?, ?, ?);"},
		{&t.updateCategory, "UPDATE [Category] SET [Name] = ?, [Order] = ?, [AdminOnly] = ? WHERE [ID] = ?;"},
		{&t.removeCategory, "DELETE FROM [Category] WHERE [ID] = ?;"},
	} {
		*stmt.Stmt, err = db.Prepare(stmt.Query)
		if err != nil {
			return errors.WithContext(fmt.Sprintf("error creating prepared statement %q: ", stmt.Query), err)
		}
	}

	trows, err := db.Query("SELECT [ID], [Name], [Category], [Price], [Duration], [Description], [Order] FROM [Treatment];")
	if err != nil {
		return errors.WithContext("error getting Treatment data: ", err)
	}

	t.treatments = make(map[uint]Treatment)

	for trows.Next() {
		var (
			tm Treatment
		)
		if err = trows.Scan(&tm.ID, &tm.Name, &tm.Category, &tm.Price, &tm.Duration, &tm.DescriptionSrc, &tm.Order); err != nil {
			return errors.WithContext("error reading Treatment row: ", err)
		}
		buildTreatmentPage(&tm)
		t.treatments[tm.ID] = tm
		buf = buf[:0]
	}
	if err = trows.Close(); err != nil {
		return errors.WithContext("error closing Treatment rows: ", err)
	}

	crows, err := db.Query("SELECT [ID], [Name], [Order], [AdminOnly] FROM [Category];")
	if err != nil {
		return errors.WithContext("error getting Category data: ", err)
	}

	t.categories = make(map[uint]Category)

	for crows.Next() {
		var cat Category
		if err = crows.Scan(&cat.ID, &cat.Name, &cat.Order, &cat.AdminOnly); err != nil {
			return errors.WithContext("error reading Category row: ", err)
		}
		t.categories[cat.ID] = cat
	}
	if err = crows.Close(); err != nil {
		return errors.WithContext("error closing Category row: ", err)
	}
	t.categoryT = template.New("")
	t.categoryT.Funcs(
		template.FuncMap{
			"price": func(amount uint) string {
				return strconv.FormatFloat(float64(amount)/100, 'f', 2, 32)
			},
		},
	)
	_, err = t.categoryT.ParseFiles(filepath.Join(*filesDir, "categories.tmpl"))
	if err != nil {
		return errors.WithContext("error parsing categories template: ", err)
	}
	t.categoryT = t.categoryT.Lookup("categories.tmpl")
	t.buildCategories()

	return nil
}

func (t *treatments) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := strconv.ParseUint(r.FormValue("id"), 10, 0)
	if err != nil {
		t.categoryPage.ServeHTTP(w, r)
		return
	}
	t.mu.RLock()
	p, ok := t.treatments[uint(id)]
	t.mu.RUnlock()
	if !ok {
		t.categoryPage.ServeHTTP(w, r)
		return
	}
	p.Description.ServeHTTP(w, r)
}

func (t *treatments) UpdateDescription(id uint, desc string) {
	buf := make(memio.Buffer, 0, 1<<20)
	bbcode.ConvertString(&buf, desc)
}

func (t *treatments) GetCategory(id uint) (Category, bool) {
	t.mu.RLock()
	c, ok := t.categories[id]
	t.mu.RUnlock()
	return c, ok
}

func (t *treatments) GetCategoryID(name string) uint {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for id, cat := range t.categories {
		if strings.EqualFold(cat.Name, name) {
			return id
		}
	}
	return 0
}

type categories []Category

func (c categories) Len() int {
	return len(c)
}

func (c categories) Less(i, j int) bool {
	if c[i].Order < c[j].Order {
		return true
	} else if c[i].Order == c[j].Order {
		return c[i].ID < c[j].ID
	}
	return false
}

func (c categories) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (t *treatments) GetCategories() []Category {
	t.mu.RLock()
	cats := make(categories, 0, len(t.categories))
	for _, cat := range t.categories {
		cats = append(cats, cat)
	}
	t.mu.RUnlock()
	sort.Sort(cats)
	return []Category(cats)
}

func (t *treatments) SetCategory(cat *Category) {
	t.mu.Lock()
	if cat.ID == 0 {
		if cat.Order == 0 {
			cat.Order = uint(len(t.categories)) + 1
		}
		res, _ := t.addCategory.Exec(cat.Name, cat.Order, cat.AdminOnly)
		id, _ := res.LastInsertId()
		cat.ID = uint(id)
	} else {
		if cat.Order == 0 {
			cat.Order = uint(len(t.categories))
		}
		t.updateCategory.Exec(cat.Name, cat.Order, cat.AdminOnly, cat.ID)
	}
	t.categories[cat.ID] = *cat
	t.buildCategories()
	t.mu.Unlock()
}

func (t *treatments) RemoveCategory(id uint) {
	t.mu.Lock()
	delete(t.categories, id)
	t.buildCategories()
	t.mu.Unlock()
	t.removeCategory.Exec(id)
}

type treatmentsS []Treatment

func (t treatmentsS) Len() int {
	return len(t)
}

func (t treatmentsS) Less(i, j int) bool {
	if t[i].Order < t[j].Order {
		return true
	} else if t[i].Order == t[j].Order {
		return t[i].ID < t[j].ID
	}
	return false
}

func (t treatmentsS) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t *treatments) GetTreatments() []Treatment {
	t.mu.RLock()
	ts := make(treatmentsS, 0, len(t.treatments))
	for _, treatment := range t.treatments {
		ts = append(ts, treatment)
	}
	t.mu.RUnlock()
	sort.Sort(ts)
	return []Treatment(ts)
}

func (t *treatments) GetTreatmentsForCategory(catID uint) []Treatment {
	t.mu.RLock()
	ts := make(treatmentsS, len(t.treatments))
	for _, treatment := range t.treatments {
		if treatment.Category == catID {
			ts = append(ts, treatment)
		}
	}
	t.mu.RUnlock()
	sort.Sort(ts)
	return []Treatment(ts)
}

func (t *treatments) SetTreatment(treatment *Treatment) {
	t.mu.Lock()
	if treatment.ID == 0 {
		if treatment.Order == 0 {
			treatment.Order = uint(len(t.treatments) + 1)
		}
		res, _ := t.addTreatment.Exec(treatment.Name, treatment.Category, treatment.Price, treatment.Duration, treatment.DescriptionSrc, treatment.Order)
		id, _ := res.LastInsertId()
		treatment.ID = uint(id)
	} else {
		if treatment.Order == 0 {
			treatment.Order = uint(len(t.treatments))
		}
		t.updateTreatment.Exec(treatment.ID, treatment.Name, treatment.Category, treatment.Price, treatment.Duration, treatment.DescriptionSrc, treatment.Order)
	}
	buildTreatmentPage(treatment)
	t.treatments[treatment.ID] = *treatment
	t.buildCategories()
	t.mu.Unlock()
}

func (t *treatments) RemoveTreatment(id uint) {
	t.removeTreatment.Exec(id)
	t.mu.Lock()
	delete(t.treatments, id)
	t.buildCategories()
	t.mu.Unlock()
}

func (t *treatments) GetTreatment(id uint) (Treatment, bool) {
	t.mu.RLock()
	treatment, ok := t.treatments[id]
	t.mu.RUnlock()
	return treatment, ok
}

func (t *treatments) GetTreatmentID(name string) uint {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for id, treatment := range t.treatments {
		if strings.EqualFold(treatment.Name, name) {
			return id
		}
	}
	return 0
}

var buf = make(memio.Buffer, 0, 1<<20)

func buildTreatmentPage(treatment *Treatment) {
	myBuf := buf
	bbcode.ConvertString(&myBuf, treatment.DescriptionSrc)
	treatment.Description = NewPageBytes("CARMEssential - "+treatment.Name, "default", template.HTML(myBuf))
}

func (t *treatments) buildCategories() {
	myBuf := buf
	cats := make(categories, 0, len(t.categories))
	for _, cat := range t.categories {
		cats = append(cats, cat)
	}
	sort.Sort(cats)
	type catTreats struct {
		Category
		Treatments []Treatment
	}
	data := make([]catTreats, len(cats))
	for n, cat := range cats {
		var treats treatmentsS
		for _, treatment := range t.treatments {
			if treatment.Category == cat.ID {
				treats = append(treats, treatment)
			}
		}
		sort.Sort(treats)
		data[n] = catTreats{
			cat,
			treats,
		}
	}
	t.categoryT.Execute(&myBuf, data)
	t.categoryPage = NewPageBytes("CARMEssential - Treatments", "treatments", template.HTML(myBuf))
}
