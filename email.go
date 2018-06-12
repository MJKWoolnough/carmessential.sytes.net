package main

import (
	"fmt"
	"net/smtp"

	"vimagination.zapto.org/errors"
)

var Email email

type email struct {
	addr, from string
	auth       smtp.Auth
}

func (e *email) Init() error {
	e.addr = Config.Get("emailSMTP")
	e.from = Config.Get("emailLogin")
	e.auth = smtp.PlainAuth(
		"",
		Config.Get("emailLogin"),
		Config.Get("emailPassword"),
		Config.Get("emailHost"),
	)
	return nil
}

func (e *email) Send(to string, msg []byte) error {
	err := smtp.SendMail(e.addr, e.auth, e.from, []string{to}, msg)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error sending email to %q: ", to), err)
	}
	return nil
}
