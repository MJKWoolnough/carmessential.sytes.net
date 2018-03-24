package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
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
	treatments           map[uint]Treatment
	categories           map[uint]Category
	sidebar, page, admin template.HTML
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

	trows, err := db.Query("SELECT [ID], [Name], [Category], [Price], [Duration], [Order] FROM [Treatment];")
	if err != nil {
		return errors.WithContext("error getting Treatment data: ", err)
	}

	t.treatments = make(map[uint]Treatment)

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
		tm.Description = NewPageBytes("CARMEssential - "+tm.Name, "", template.HTML(buf))
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

	t.categories = make(map[uint]Category)

	for crows.Next() {
		var cat Category
		if err = crows.Scan(&cat.ID, &cat.Name, &cat.Order, &cat.AdminOnly); err != nil {
			return errors.WithContext("error reading Category row: ", err)
		}
		t.categories[cat.ID] = &cat
	}
	if err = crows.Close(); err != nil {
		return errors.WithContext("error closing Category row: ", err)
	}

	return nil
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

func (t *treatments) GetCategoryID(name string) uint {
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
	}
	return c[i].ID < c[j].ID
}

func (t *treatments) GetCategories() []Category {
	cats := make(categories, 0, len(t.categories))
	for _, o := range t.categoryOrder {
		cats = append(cats, *t.categories[o])
	}
	sort.Sort(cats)
	return []Category(cats)
}

func (t *treatments) SetCategory(cat *Category) {
	if cat.ID == 0 {
		if cat.Order == 0 {
			cat.Order = len(t.categories) + 1
		}
		id, _ := t.addCategory.Exec(cat.Name, cat.Order, cat.AdminOnly).LastInsertId()
		cat.ID = uint(id)
		t.categories[uint(id)] = *cat
	} else {
		if cat.Order == 0 {
			cat.Order = len(t.categories)
		}
		t.updateCategory.Exec(cat.Name, cat.Order, cat.AdminOnly, cat.ID)
		t.categories[cat.ID] = *cat
	}
}
