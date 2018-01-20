package main

import "io"

const (
	VoucherTreatment = iota
	VoucherMoney
)

type Vouchera struct {
	VoucherType int
	Amount      int
	Treatment   string
}

func (v *Vouchera) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}
