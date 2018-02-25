package main

import "net/smtp"

var Email email

type email struct {
	addr string
	auth smtp.Auth
	from string
}

func (e *email) init(addr string, auth smtp.Auth, from string) {
	e.addr = addr
	e.auth = auth
	e.from = from
}

func (e *email) Send(to string, msg []byte) error {
	return smtp.SendMail(e.addr, e.auth, e.from, []string{to}, msg)
}
