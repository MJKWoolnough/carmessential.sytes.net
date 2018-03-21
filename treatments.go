package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/MJKWoolnough/errors"
	"github.com/MJKWoolnough/memio"
)

var Treatments treatments

type Treatment struct {
	ID          uint
	Name        string
	Category    uint
	Price       uint
	Duration    time.Duration
	Order       uint
	Description PageBytes
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

	sync.RWMutex
	treatments                    treatmentMap
	categories                    categoryMap
	treatmentOrder, categoryOrder []uint
	sidebar, page, admin          template.HTML
}

type treatmentMap map[uint]*Treatment

func (t treatmentMap) order(id uint) uint {
	return t[id].Order
}

type categoryMap map[uint]*Category

func (c categoryMap) order(id uint) uint {
	return c[id].Order
}

func (t *treatments) init(db *sql.DB) error {
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
		{&t.addCategory, "INSERT INTO [Category]([Name]) VALUES (?);"},
		{&t.updateCategory, "UPDATE [Category] SET [Name] = ?, [Order] = ?, [AdminOnly] = ? WHERE [ID] = ?;"},
		{&t.removeCategory, "DELETE FROM [Category] WHERE [ID] = ?;"},
	} {
		*stmt.Stmt, err = db.Prepare(stmt.Query)
		if err != nil {
			return errors.WithContext(fmt.Sprintf("error creating prepared statement %q: ", stmt.Query), err)
		}
	}

	trows, err := db.Query("SELECT [ID], [Name], [Category], [Price], [Duration], [Order] FROM [Treatment];")
	if err != nil {
		return errors.WithContext("error getting Treatment data: ", err)
	}

	t.treatments = make(treatmentMap)

	buf := make(memio.Buffer, 0, 1<<20)
	for trows.Next() {
		var (
			tm          Treatment
			description string
		)
		if err = trows.Scan(&tm.ID, &tm.Name, &tm.Category, &tm.Price, &tm.Duration, &description, &tm.Order); err != nil {
			return errors.WithContext("error reading Treatment row: ", err)
		}
		bbcode.ConvertString(&buf, description)
		//tm.Description = template.HTML(buf)
		t.treatmentOrder = append(t.treatmentOrder, tm.ID)
		t.treatments[tm.ID] = &tm
		buf = buf[:0]
	}
	if err = trows.Close(); err != nil {
		return errors.WithContext("error closing Treatment rows: ", err)
	}

	crows, err := db.Query("SELECT [ID], [Name], [Order], [AdminOnly] FROM [Category];")
	if err != nil {
		return errors.WithContext("error getting Category data: ", err)
	}

	t.categories = make(categoryMap)

	for crows.Next() {
		var cat Category
		if err = crows.Scan(&cat.ID, &cat.Name, &cat.Order, &cat.AdminOnly); err != nil {
			return errors.WithContext("error reading Category row: ", err)
		}
		t.categoryOrder = append(t.categoryOrder, cat.ID)
		t.categories[cat.ID] = &cat
	}
	if err = crows.Close(); err != nil {
		return errors.WithContext("error closing Category row: ", err)
	}

	t.generateHTML()

	return nil
}

type treatmentOrder interface {
	order(uint) uint
}

type treatmentSorter struct {
	list []uint
	treatmentOrder
}

func (t treatmentSorter) Len() int {
	return len(t.list)
}

func (t treatmentSorter) Less(i, j int) bool {
	return t.treatmentOrder.order(t.list[i]) < t.treatmentOrder.order(t.list[j])
}

func (t treatmentSorter) Swap(i, j int) {
	t.list[i], t.list[j] = t.list[j], t.list[i]
}

func (t *treatments) generateHTML() {
	sort.Sort(treatmentSorter{
		list:           t.treatmentOrder,
		treatmentOrder: t.treatments,
	})
	sort.Sort(treatmentSorter{
		list:           t.categoryOrder,
		treatmentOrder: t.categories,
	})
	// TODO: generate sidebar, page and admin
}

func (t *treatments) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := strconv.ParseUint(r.FormValue("id"), 10, 0)
	if err != nil {
		t.ServeCategories(w, r)
		return
	}
	t.RLock()
	p, ok := t.treatments[uint(id)]
	_ = p
	t.RUnlock()
	if !ok {
		t.ServeCategories(w, r)
		return
	}

}

func (t *treatments) ServeCategories(w http.ResponseWriter, r *http.Request) {

}

func (t *treatments) UpdateDescription(id uint, desc string) {
	buf := make(memio.Buffer, 0, 1<<20)
	bbcode.ConvertString(&buf, desc)
}

func (t *treatments) GetCategory(id uint) (Category, bool) {
	c, ok := t.categories[id]
	if ok {
		return *c, ok
	}
	return Category{}, false
}

func (t *treatments) GetCategories() []Category {
	cats := make([]Category, 0, len(t.categories))
	for _, o := range t.categoryOrder {
		cats = append(cats, *t.categories[o])
	}
	return cats
}
