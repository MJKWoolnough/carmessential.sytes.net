package main

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/MJKWoolnough/errors"
	_ "github.com/mattn/go-sqlite3"
)

var DB db

type db struct {
	sync.Mutex
	*sql.DB
}

func (db *db) init() error {
	database, err := sql.Open("sqlite3", *databaseFile)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error opening database file %q: ", *databaseFile), err)
	}
	if err = Config.init(database); err != nil {
		return errors.WithContext("error initialising Config: ", err)
	}
	if err = Treatments.init(database); err != nil {
		return errors.WithContext("error initialising Treatments: ", err)
	}
	if err = Users.init(database); err != nil {
		return errors.WithContext("error initialising Users: ", err)
	}
	db.DB = database
	return nil
}

func (db *db) Close() error {
	return db.DB.Close()
}
