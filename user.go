package main

import (
	"encoding/base64"
	"html/template"
	"net/http"
	"time"

	"github.com/MJKWoolnough/authenticate"
)

var User user

type user struct {
	loginT, registerT, emailT *template.Template
	from                      string
	registerCodec             *authenticate.Codec
}

func (u *user) init(login, register, email, from string, registerKey []byte) error {
	var err error
	u.loginT, err = template.ParseFiles(login)
	if err != nil {
		return err
	}
	u.registerT, err = template.ParseFiles(register)
	if err != nil {
		return err
	}
	u.emailT, err = template.ParseFiles(email)
	if err != nil {
		return err
	}
	u.registerCodec, err = authenticate.NewCodec(registerKey, time.Hour*24)
	return err
}

func (u *user) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch Session.GetLogin(r) {
	case 0:
		http.Redirect(w, r, "/login.html", http.StatusFound)
		return
	case 1:
		http.Redirect(w, r, "/admin/", http.StatusFound)
		return
	}
	Pages.WriteHeader(w, r, PageHeader{
		Title:       []byte("CARMEssential - User Area"),
		Style:       []byte("user"),
		WriteBasket: true,
	})
	w.Write([]byte("USER"))
	Pages.WriteFooter(w)
}

var registerPH = PageHeader{
	Title:       []byte("CARMEssential - Register"),
	Style:       []byte("user"),
	WriteBasket: true,
}

func isValidPassword(password string) bool {
	return true
}

func isValidEmail(email string) bool {
	return true
}

func (u *user) Register(w http.ResponseWriter, r *http.Request) {
	if Session.GetLogin(r) > 0 {
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	r.ParseForm()
	var form struct {
		Email, Code, Error, From string
		Stage                    int
	}
	form.Code = r.Form.Get("code")
	if form.Code != "" {
		code, err := base64.StdEncoding.DecodeString(form.Code)
		if err == nil {
			email, err := u.registerCodec.Decode(code, nil)
			if err != nil {
				form.Code = ""
			} else {
				form.Email = string(email)
				form.Stage = 2
				pass := r.Form.Get("password")
				if pass != "" {
					if pass == r.Form.Get("confirmPassword") {
						if isValidPassword(pass) {
							// set db
							http.Redirect(w, r, "/user/", http.StatusFound)
							return
						} else {
							form.Error = "Invalid Password"
						}
					} else {
						form.Error = "Passwords do not match"
					}
				}
			}
		}
	} else {
		form.Email = r.Form.Get("email")
		if form.Email != "" {
			if isValidEmail(form.Email) {
				form.Code = base64.StdEncoding.EncodeToString(u.registerCodec.Encode([]byte(form.Email), make([]byte, 0, len(form.Email)+u.registerCodec.Overhead())))
				form.From = u.from
				Email.Send(form.Email, emailTemplate{u.emailT, form})
				//u.emailT.Execute(e, form)

				// send email
				form.Stage = 1
			} else {
				form.Error = "Invalid Email Address"
			}
		}
	}
	Pages.WriteHeader(w, r, registerPH)
	u.registerT.Execute(w, &form)
	Pages.WriteFooter(w)
}

var loginPH = PageHeader{
	Title:       []byte("CARMEssential - Login"),
	Style:       []byte("user"),
	WriteBasket: true,
}

func (u *user) Login(w http.ResponseWriter, r *http.Request) {
	if Session.GetLogin(r) > 0 {
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	Pages.WriteHeader(w, r, loginPH)
	w.Write([]byte("LOGIN"))
	Pages.WriteFooter(w)
}

func (u *user) Logout(w http.ResponseWriter, r *http.Request) {
	Session.ClearLogin(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
