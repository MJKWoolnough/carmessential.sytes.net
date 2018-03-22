package main

import (
	"io"
	"time"

	"github.com/MJKWoolnough/errors"
)

func BasketInit() error {
	if err := Pages.RegisterTemplate("basket.tmpl"); err != nil {
		return errors.WithContext("error registering basket template: ", err)
	}
	// register types
	return nil
}

type Basket struct {
	Items     []Item
	Vouchers  map[string]Voucher
	Discounts map[string]Discount
}

func (b *Basket) IsEmpty() bool {
	return b == nil || len(b.Items) == 0
}

func (b *Basket) Print(w io.Writer) string {
	return ""
}

func (b *Basket) SubTotal() uint {
	var total uint
	for _, i := range b.Items {
		total += i.Price()
	}
	return total
}

type Discount interface {
	Process([]Item) uint
}

type Voucher interface {
	Process([]Item) uint
}

type Item interface {
	Name() string
	Price() uint
}

type Qty interface {
	Qty() uint32
	QtyAdd(uint32) uint32
	QtySub(uint32) uint32
}

type Quantity uint32

func (q Quantity) Qty() uint32 {
	return uint32(q)
}

func (q *Quantity) QtyAdd(diff uint32) uint32 {
	*q += Quantity(diff)
	return uint32(*q)
}

func (q *Quantity) QtySub(diff uint32) uint32 {
	if *q <= Quantity(diff) {
		*q = 0
		return 0
	}
	*q -= Quantity(diff)
	return uint32(*q)
}

type Product struct {
	ID    int
	price uint
	Quantity
}

func (p *Product) Price() uint {
	return p.price * uint(p.Quantity)
}

type Service struct {
	ID    int
	Time  time.Time
	price uint
}

func (s *Service) Price() uint {
	return s.price
}
