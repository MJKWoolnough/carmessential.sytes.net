package main

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mxk/go-sqlite/sqlite3"
)

var db database

type database struct {
	mu sync.Mutex
	db *sql.DB
}

func SetupDatabase(filename string) error {
	return nil
}

type User struct{}

func (d *database) CreateUser(u *User) error {
	return nil
}

func (d *database) LoginUser(u *user) error {
	return nil
}

func (d *database) UpdateUser(u *user) error {
	return nil
}

func (d *database) GetUser(u *user) error {
	return nil
}

func (d *database) CreateTreatment(t *treatment) error {
	return nil
}

func (d *database) UpdateTreatment(t *treatment) error {
	return nil
}

func (d *database) GetTreatment(t *treatment) error {
	return nil
}

func (d *database) GetTreatments(tid int) ([]treatment, error) {
	return nil, nil
}

func (d *database) CreateProduct(p *product) error {
	return nil
}

func (d *database) UpdateProduct(p *product) error {
	return nil
}

func (d *database) GetProduct(p *product) error {
	return nil
}

func (d *database) GetProducts(pid int) ([]product, error) {
	return nil, nil
}

func (d *database) CreateBooking(b *booking) error {
	return nil
}

func (d *database) GetBooking(b *booking) error {
	return nil
}

func (d *database) UpdateBooking(b *booking) error {
	return nil
}

func (d *database) GetBookings(from, to time.Time) ([]booking, error) {
	return nil, nil
}

func (d *database) CreateVoucher(v *voucher) error {
	return nil
}

func (d *database) GetVoucher(v *voucher) error {
	return nil
}

func (d *database) UseVoucher(v *voucher) error {
	return nil
}
