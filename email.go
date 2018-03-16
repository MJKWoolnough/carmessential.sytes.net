package main

import (
	"fmt"
	"net/smtp"

	"github.com/MJKWoolnough/errors"
)

var Email email

type email struct {
	addr string
	auth smtp.Auth
	from string
}

func (e *email) init(addr, from string, auth smtp.Auth) error {
	e.addr = addr
	e.auth = auth
	e.from = from
	return nil
}

func (e *email) Send(to string, msg []byte) error {
	err := smtp.SendMail(e.addr, e.auth, e.from, []string{to}, msg)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error sending email to %q: ", to), err)
	}
	return nil
}
