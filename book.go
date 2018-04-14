package main

import (
	"context"
	"database/sql"
	"net/http"
	"sort"
	"strconv"
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

func timeToDate(t time.Time) uint {
	y, m, d := time.Date()
	return uint(y)*10000 + uint(m)*100 + uint(d)
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
		date := timeToDate(booking.Time)
		b.bookings[date] = append(b.bookings[date], booking)
	}

	for _, bookings := range b.bookings {
		sort.Sort(bookings)
	}

	if err = Pages.RegisterTemplate("booking.tmpl"); err != nil {
		return errors.WithContext("error registering booking template: ", err)
	}
	if err = Pages.RegisterTemplate("bookingConfirmation.tmpl"); err != nil {
		return errors.WithContext("error registering booking confirmation template: ", err)
	}
	return nil
}

type DayData struct {
	Date     time.Time
	Bookings bookingList
}

func (b *bookings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	treatmentID, _ := strconv.ParseUint(r.PostForm.Get("id"), 10, 32)
	treatment, exists := Treatments.GetTreatment(uint(treatmentID))
	if !exists {
		http.Redirect(w, r, "/")
		return
	}
	if timeStr := r.PostForm.Get("time"); timeStr != "" {
		timeSec, err := strconv.ParseUint(timeStr, 10, 64)
		if err != nil {
			http.Error(w, "error", http.StatusNotAcceptable)
			return
		}
		// validate time
		// write confirmation page & book
	}
	page, _ := strconv.ParseUint(r.PostForm.Get("page"), 10, 32)
	today := time.Now()
	days := make([]DayData, 7)
	for i := 0; i < 7; i++ {
		day := today.Add(time.Hour * 24 * time.Duration(uint64(i)+page*7))
		days[i] = DayData{
			Date:     day,
			Bookings: b.bookings[timeToDate(day)],
		}
	}
	uid := Session.GetLogin(r)
	basket := Session.LoadBasket(r)
	r = r.WithContext(context.WithValue(context.WithValue(r.Context(), "userID", uid), "basket", basket))
	Pages.Write(w, r, "booking.tmpl", struct {
		Admin bool
		*Basket
		Page uint
		Treatment
		DayData
	}{
		Users.IsAdmin(uid),
		basket,
		page,
		treatment,
		days,
	})
}
