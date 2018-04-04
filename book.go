package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/MJKWoolnough/errors"
)

var Bookings bookings

type bookings struct {
	bookings                  map[time.Time]bookingList
	addBooking, updateBooking *sql.Stmt
}

type Booking struct {
	ID   uint
	Time time.Time
}

type bookingList []Booking

func (b bookingList) Len() int {
	return len(b)
}

func (b bookingList) Less(i, j int) bool {
	return !b[j].Time.After(b[i].Time)
}

func (b bookingList) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b *bookings) Init(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS [Booking]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [ID] INTEGER NOT NULL, [Treatment] INTEGER NOT NULL, [User] INTEGER NOT NULL DEFAULT 0, [Name] STRING NOT NULL);")
	if err != nil {
		return errors.WithContext("error creating Booking table: ", err)
	}

	if err = Pages.RegisterTemplate("booking.tmpl"); err != nil {
		return errors.WithContext("error registering booking template: ", err)
	}
	return nil
}

func (b *bookings) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
