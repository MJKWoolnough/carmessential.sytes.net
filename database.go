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
	var err error
	if db.DB, err = sql.Open("sqlite3", filename); err != nil {
		return err
	}
	if err = users.init(); err != nil {
		return err
	}
	return nil
}
