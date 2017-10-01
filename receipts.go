package main

import (
	"io"
	"time"
)

const (
	ReceiptItemTreatment = iota
	ReceiptItemProduct
)

type ReceiptItem struct {
	Name     string
	Type     int
	Date     time.Time
	Quantity int
	Total    int
}

type ReceiptDiscount struct {
	Name  string
	Total int
}

type Receipt struct {
	Items     []ReceiptItem
	Postage   int
	Discounts []ReceiptDiscount
}

func (r *Receipt) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}
