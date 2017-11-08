package main

import (
	"bytes"
	"database/sql"
	"html/template"
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

type Treatments struct {
	treatments                                     map[uint]*Treatment
	order                                          []uint
	categories                                     map[uint]string
	addTreatment, updateTreatment, removeTreatment *sql.Stmt
	addCategory, updateCategory, removeCategory    *sql.Stmt
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
		t.order = append(t.order, tm.ID)
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
	for crows.Next() {
		var (
			id   uint
			name string
		)
		if err = crows.Scan(&id, &name); err != nil {
			return err
		}
		t.categories[id] = name
	}
	if err = crows.Close(); err != nil {
		return err
	}
	return nil
}
