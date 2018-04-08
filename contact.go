package main

import (
	"fmt"
	"net/http"

	"github.com/MJKWoolnough/errors"
	"github.com/MJKWoolnough/form"
)

var Contact contact

type contact struct{}

func (c *contact) Init() error {
	if err := Pages.RegisterTemplate("contact.tmpl"); err != nil {
		return errors.WithContext("error registering contact template: ", err)
	}
	return nil
}

type contactValues struct {
	Name, Email, Phone, Subject, Message string
	Errors                               form.Errors
	Done                                 bool
}

func (v *contactValues) ParserList() form.ParserList {
	return form.ParserList{
		"name":    form.RequiredString{&v.Name},
		"email":   form.RequiredString{&v.Email},
		"phone":   form.String{&v.Phone},
		"subject": form.String{&v.Subject},
		"message": form.String{&v.Message},
	}
}

func (c *contact) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var v contactValues
	if r.Method == http.MethodPost {
		r.ParseForm()
		if r.Form.Get("submit") != "" {
			err := form.Parse(&v, r.PostForm)
			if err == nil {
				to := Config.Get("contactTo")
				if err = Email.Send(to, []byte(fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: Message Received\r\n\r\nName: %s\nEmail: %s\nPhone: %s\nSubject: %s\nMessage: %s", to, to, v.Email, v.Phone, v.Subject, v.Message))); err != nil {
					v.Errors = form.Errors{"send": err}
				}
				v.Done = true
			} else {
				v.Errors = err.(form.Errors)
			}
		}
	}
	Pages.Write(w, r, "contact.tmpl", &v)
}
