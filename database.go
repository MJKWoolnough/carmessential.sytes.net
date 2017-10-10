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
	DBTotalStatements
)

type database struct {
	db *sql.DB

	mu         sync.Mutex
	statements [DBTotalStatements]*sql.Stmt
}

func (d *database) Init(filename string) error {
	var err error
	d.db, err = sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}
	for _, table := range [...]string{
		"CREATE TABLE IF NOT EXIST [User]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Email] TEXT NOT NULL, [Password] TEXT NOT NULL);",
		"CREATE TABLE IF NOT EXIST [TreatmentGroup]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL);",
		"CREATE TABLE IF NOT EXIST [Treatment]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Description] TEXT NOT NULL, [Group] INTEGER NOT NULL, [Duration] INTEGER NOT NULL, [Price] INTEGER NOT NULL);",
		"CREATE TABLE IF NOT EXIST [Products]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Description] TEXT NOT NULL, [Price] INTEGER NOT NULL);",
		"CREATE TABLE IF NOT EXIST [Order]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Basket] TEXT NOT NULL, [Completed] INTEGER NOT NULL, [Status] INTEGER NOT NULL);",
		"CREATE TABLE IF NOT EXIST [Voucher]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Code] TEXT NOT NULL, [Name] TEXT NOT NULL, [Expiry] INTEGER NOT NULL, [OrderID] INTEGER NOT NULL, [Type] INTEGER NOT NULL, [Value] INTEGER NOT NULL, [Valid] BOOLEAN DEFAULT 0 NOT NULL CHECK ([Valid] IN (0,1));",
		"CREATE TABLE IF NOT EXITS [Booking]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Order] INTEGER NOT NULL, [Treatment] INTEGER NOT NULL, [Time] INTEGER NOT NULL, [Name] TEXT NOT NULL, [EmailAddress] TEXT NOT NULL, [PhoneNumber] TEXT NOT NULL);",
	} {
		_, err = d.db.Exec(table)
		if err != nil {
			return err
		}
	}

	for n, ps := range [TotalStatements]string{
		"INSERT INTO [User]([Name], [Email], [Password]) VALUES (?, ?, ?);",
	} {
		d.statements[n], err = d.db.Prepare(ps)
		if err != nil {
			return err
		}
	}

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
