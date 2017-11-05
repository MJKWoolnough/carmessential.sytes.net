package main

import "time"

type Basket struct {
	Items    []Item
	Vouchers map[string]Voucher
}

type Item interface {
}

type Qty interface {
	Qty() uint32
	QtyAdd(uint32) uint32
	QtySub(uint32) uint32
}

type Quantity uint32

func (q Quantity) Qty() uint32 {
	return q
}

func (q *Quantity) QtyAdd(diff uint32) uint32 {
	*q += diff
	return *q
}

func (q *Quantity) QtySub(diff uint32) uint32 {
	if *q <= diff {
		*q = 0
		return 0
	}
	*q -= diff
	return *q
}

type Product struct {
	ID int
	Quantity
}

type Service struct {
	ID   int
	Time time.Time
}
