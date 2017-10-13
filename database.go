package main

import (
	"database/sql"
	"sync"
	"time"

	_ "github.com/mxk/go-sqlite/sqlite3"
)

var db database

const (
	DBCreateUser = iota
	DBCreateTreatmentGroup
	DBCreateTreatment
	DBCreateProduct
	DBCreateOrder
	DBCreateVoucher
	DBCreateBooking
	DBCreateOrderedProduct
	DBTotalStatements
)

type database struct {
	mu sync.Mutex
	db *sql.DB
}

func (d *database) init(filename string) error {
	var err error
	if d.db, err = sql.Open("sqlite3", filename); err != nil {
		return err
	}
	if err = users.init(d); err != nil {
		return err
	}
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
