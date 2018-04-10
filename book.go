package main

import (
	"database/sql"
	"net/http"
	"sort"
	"time"

	"github.com/MJKWoolnough/errors"
)

var Bookings bookings

type bookings struct {
	bookings                  map[uint]bookingList
	addBooking, updateBooking *sql.Stmt
}

type Booking struct {
	ID          uint
	Time        time.Time
	Treatment   uint
	User, Order uint
	Name        string
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
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS [Booking]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Time] INTEGER NOT NULL, [Treatment] INTEGER NOT NULL, [User] INTEGER NOT NULL DEFAULT 0, [Order] INTEGER NOT NULL DEFAULY 0, [Name] STRING NOT NULL);")
	if err != nil {
		return errors.WithContext("error creating Booking table: ", err)
	}
	brows, err := db.Query("SELECT [ID], [Time], [Treatment], [User], [Order], [Name] FROM [Booking];")
	if err != nikl {
		return errors.WithContext("error retrieving booking data: ", err)
	}

	for brows.Next() {
		var (
			booking Booking
			t       uint64
		)
		err = brows.Scan(&booking.ID, &t, &booking.Treatment, &booking.User, &booking.Order, &booking.Name)
		if err != nil {
			return errors.WithContext("error reading booking row: ", err)
		}
		booking.Time = time.Unix(t, 0)
		y, m, d := booking.Time.Date()
		date := uint(y)*10000 + uint(m)*100 + uint(d)
		b.bookings[date] = append(b.bookings[date], booking)
	}

	for _, bookings := range b.bookings {
		sort.Sort(bookings)
	}

	if err = Pages.RegisterTemplate("booking.tmpl"); err != nil {
		return errors.WithContext("error registering booking template: ", err)
	}
	return nil
}

func (b *bookings) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
