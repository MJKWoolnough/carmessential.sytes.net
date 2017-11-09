package main

import (
	"bytes"
	"database/sql"
	"html/template"
	"sort"
	"sync"
	"time"
)

var treatments Treatments

type Treatment struct {
	ID          uint
	Name        string
	Category    uint
	Price       uint
	Duration    time.Duration
	Order       uint
	Description template.HTML
}

type Category struct {
	ID    uint
	Name  string
	Order uint
}

type Treatments struct {
	addTreatment, updateTreatment, removeTreatment *sql.Stmt
	addCategory, updateCategory, removeCategory    *sql.Stmt

	sync.Mutex
	treatments                    treatmentMap
	categories                    categoryMap
	treatmentOrder, categoryOrder []uint
}

type treatmentMap map[uint]*Treatment

func (t treatmentMap) order(id uint) uint {
	return t[id].Order
}

type categoryMap map[uint]*Category

func (c categoryMap) order(id uint) uint {
	return c[id].Order
}

func (t *Treatments) init() error {
	db.Lock()
	defer db.Unlock()

	_, err := db.Exec("CREATE TABLE IF NOT EXIST [Treatment]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Category] INTEGER NOT NULL, [Price] INTEGER NOT NULL, [Duration] INTEGER NOT NULL, [Description] TEXT NOT NULL, [Order] INTEGER NOT NULL);")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXIST [Category]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL);")
	if err != nil {
		return err
	}

	for stmt, query := range map[**sql.Stmt]string{
		&t.addTreatment:    "INSERT INTO [Treatment]([Name], [Category], [Price], [Duration], [Description], [Order]) VALUES (?, ?, ?, ?, ?, ?);",
		&t.updateTreatment: "UPDATE [Treatment] SET [Name] = ?, [Category] = ?, [Price] = ?, [Duration] = ?, [Description] = ?, [Order] = ? WHERE [ID] = ?;",
		&t.removeTreatment: "DELETE FROM [Treatment] WHERE [ID] = ?;",
		&t.addCategory:     "INSERT INTO [Category]([Name]) VALUES (?);",
		&t.updateCategory:  "UPDATE [Category] SET [Name] = ? WHERE [ID] = ?;",
		&t.removeCategory:  "DELETE FROM [Category] WHERE [ID] = ?;",
	} {
		*stmt, err = db.Prepare(query)
		if err != nil {
			return err
		}
	}

	trows, err := db.Query("SELECT [ID], [Name], [Category], [Price], [Duration], [Order] FROM [Treatment];")
	if err != nil {
		return err
	}

	t.treatments = make(treatmentMap)

	var w bytes.Buffer
	for trows.Next() {
		var (
			tm          Treatment
			description string
		)
		if err = trows.Scan(&tm.ID, &tm.Name, &tm.Category, &tm.Price, &tm.Duration, &description, &tm.Order); err != nil {
			return err
		}
		bbcode.ConvertString(&w, description)
		tm.Description = template.HTML(w.String())
		t.treatmentOrder = append(t.treatmentOrder, tm.ID)
		t.treatments[tm.ID] = &tm
		w.Reset()
	}
	if err = trows.Close(); err != nil {
		return err
	}

	crows, err := db.Query("SELECT [ID], [Name] FROM [Category];")
	if err != nil {
		return err

	}

	t.categories = make(categoryMap)

	for crows.Next() {
		var cat Category
		if err = crows.Scan(&cat.ID, &cat.Name, &cat.Order); err != nil {
			return err
		}
		t.categoryOrder = append(t.categoryOrder, cat.ID)
		t.categories[cat.ID] = &cat
	}
	if err = crows.Close(); err != nil {
		return err
	}

	sort.Sort(treatmentSorter{
		list:           t.treatmentOrder,
		treatmentOrder: t.treatments,
	})
	sort.Sort(treatmentSorter{
		list:           t.categoryOrder,
		treatmentOrder: t.categories,
	})

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
