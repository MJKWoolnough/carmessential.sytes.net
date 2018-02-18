package main

import (
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var DB db

type db struct {
	sync.Mutex
	*sql.DB
}

func (db *db) init(filename string) error {
	database, err := sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}
	if err = treatments.init(database); err != nil {
		return err
	}
	if err = users.init(database); err != nil {
		return err
	}
	db.DB = database
	return nil
}

func (db *db) Close() error {
	return db.DB.Close()
}
