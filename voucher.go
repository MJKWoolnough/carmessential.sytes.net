package main

import "io"

const (
	VoucherTreatment = iota
	VoucherMoney
)

type Voucher struct {
	VoucherType int
	Amount      int
	Treatment   string
}

func (v *Voucher) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}
