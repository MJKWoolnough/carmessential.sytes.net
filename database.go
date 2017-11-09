package main

import (
	"database/sql"
	"sync"

	_ "github.com/mxk/go-sqlite/sqlite3"
)

var db struct {
	sync.Mutex
	*sql.DB
}

func dbInit(filename string) error {
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
