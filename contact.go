package main

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/MJKWoolnough/errors"
	"github.com/MJKWoolnough/form"
	"github.com/MJKWoolnough/memio"
)

var Contact contact

type contact struct {
	emailT *template.Template
}

func (c *contact) Init() error {
	err := Pages.RegisterTemplate("contact.tmpl")
	if err != nil {
		return errors.WithContext("error registering contact page template: ", err)
	}
	c.emailT, err = template.ParseFiles(filepath.Join(*filesDir, "contactEmail.tmpl"))
	if err != nil {
		return errors.WithContext("error parsing contact email template: ", err)
	}
	return nil
}

type contactValues struct {
	To, From                             string
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
				v.To = to
				v.From = to
				var buf memio.Buffer
				c.emailT.Execute(&buf, &v)
				if err = Email.Send(to, buf); err != nil {
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
