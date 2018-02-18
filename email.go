package main

import (
	"io"
	"net/smtp"
	"time"
)

var Email email

type email struct {
	host    string
	auth    smtp.Auth
	from    string
	timeout time.Duration

	send chan sendEmail
}

func (e *email) init(host, from string, auth smtp.Auth, timeout time.Duration) {
	e.host = host
	e.from = from
	e.auth = auth
	e.timeout = timeout
	e.send = make(chan sendEmail)
}

type sendEmail struct {
	to   string
	data io.WriterTo
}

func (e *email) Send(to string, data io.WriterTo) {
	e.send <- sendEmail{to, data}
}

// runs in its own goroutine
func (e *email) run() {
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	var (
		client *smtp.Client
		err    error
	)
	for {
		select {
		case timer.C:
			client.Quit()
			client.Close()
			client = nil
		case se := <-e.send:
			if client != nil && client.Noop() != nil {
				client.Close()
				client = nil
			}
			if client == nil {
				client = smtp.Dial(e.host)
				client.Auth(e.auth)
			}

			err = client.Mail(e.from)
			if err != nil {
				client.Reset()
				//TODO:handle
			}

			err = client.Rcpt(se.to)
			if err != nil {
				client.Reset()
				//TODO:handle
			}

			wc, err := client.Data()
			if err != nil {
				client.Reset()
				//TODO:handle
			}
			_, err = se.data.WriteTo(wc)
			if err != nil {
				client.Reset()
				//TODO:handle
			}
			wc.Close()

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(e.timeout)
		}
	}
}
