package main

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mxk/go-sqlite/sqlite3"
)

var db database

type database struct {
	db *sql.DB

	mu         sync.Mutex
	statements []*sql.Stmt
}

func SetupDatabase(filename string) error {
	return nil
}

type User struct{}

func (d *database) CreateUser(u *User) error {
	return nil
}

func (d *database) LoginUser(u *User) error {
	return nil
}

func (d *database) UpdateUser(u *User) error {
	return nil
}

func (d *database) GetUser(u *User) error {
	return nil
}

func (d *database) CreateTreatment(t *Treatment) error {
	return nil
}

func (d *database) UpdateTreatment(t *Treatment) error {
	return nil
}

func (d *database) GetTreatment(t *Treatment) error {
	return nil
}

func (d *database) GetTreatments(tid int) ([]Treatment, error) {
	return nil, nil
}

func (d *database) CreateProduct(p *Product) error {
	return nil
}

func (d *database) UpdateProduct(p *Product) error {
	return nil
}

func (d *database) GetProduct(p *Product) error {
	return nil
}

func (d *database) GetProducts(pid int) ([]Product, error) {
	return nil, nil
}

func (d *database) CreateBooking(b *Booking) error {
	return nil
}

func (d *database) GetBooking(b *Booking) error {
	return nil
}

func (d *database) UpdateBooking(b *Booking) error {
	return nil
}

func (d *database) GetBookings(from, to time.Time) ([]Booking, error) {
	return nil, nil
}

func (d *database) CreateVoucher(v *Voucher) error {
	return nil
}

func (d *database) GetVoucher(v *Voucher) error {
	return nil
}

func (d *database) UseVoucher(v *Voucher) error {
	return nil
}
