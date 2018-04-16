package main

import (
	"encoding/base64"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/MJKWoolnough/authenticate"
	"github.com/MJKWoolnough/errors"
	"github.com/MJKWoolnough/memio"
	vpages "github.com/MJKWoolnough/pages"
)

var User user

type user struct {
	emailT        *template.Template
	from          string
	registerCodec *authenticate.Codec

	tempUserPage *vpages.Bytes
}

func (u *user) Init() error {
	u.from = Config.Get("emailFrom")
	err := Pages.RegisterTemplate("login.tmpl")
	if err != nil {
		return errors.WithContext("error registering Login template: ", err)
	}
	err = Pages.RegisterTemplate("register.tmpl")
	if err != nil {
		return errors.WithContext("error registering Register template: ", err)
	}
	u.emailT, err = template.ParseFiles(filepath.Join(*filesDir, "registrationEmail.tmpl"))
	if err != nil {
		return errors.WithContext("error registering Email template: ", err)
	}
	u.registerCodec, err = authenticate.NewCodec([]byte(Config.Get("registrationKey")), time.Hour*24)
	if err != nil {
		return errors.WithContext("error creating registration authenticator: ", err)
	}

	u.tempUserPage = Pages.Bytes("CARM Essential - User", "default", "USER INDEX")

	return nil
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
	u.tempUserPage.ServeHTTP(w, r)
}

func isValidPassword(password string) bool {
	return len(password) > 0
}

func isValidEmail(email string) bool {
	return len(email) > 0
}

func isValidPhone(phone string) bool {
	return true
}

func (u *user) Register(w http.ResponseWriter, r *http.Request) {
	if Session.GetLogin(r) > 0 {
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	r.ParseForm()
	var form struct {
		Email, Code, Error, From, Name, NameError, Phone, PhoneError string
		Stage                                                        int
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
				if id, _ := Users.UserID(form.Email); id > 0 {
					form.Error = "Email address already registered"
				} else {
					form.Stage = 2
					if _, ok := r.Form["password"]; ok {
						form.Name = r.Form.Get("name")
						if form.Name == "" {
							form.NameError = "Need your Name"
						}
						form.Phone = r.Form.Get("phone")
						if !isValidPhone(form.Phone) {
							form.PhoneError = "Invalid Phone Number"
						}
						if pass := r.Form.Get("password"); pass == r.Form.Get("confirmPassword") {
							if isValidPassword(pass) {
								id, _ := Users.CreateUser(form.Name, form.Email, pass, form.Phone)
								Session.SetLogin(w, id)
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
		}
	} else {
		form.Email = r.Form.Get("email")
		if id, _ := Users.UserID(form.Email); id > 0 {
			form.Error = "Email address already registered"
		} else if isValidEmail(form.Email) {
			form.Code = base64.StdEncoding.EncodeToString(u.registerCodec.Encode([]byte(form.Email), make([]byte, 0, len(form.Email)+u.registerCodec.Overhead())))
			form.From = u.from
			var msg memio.Buffer
			u.emailT.Execute(&msg, form)
			Email.Send(form.Email, msg)
			form.Stage = 1
		} else if _, ok := r.Form["email"]; ok {
			form.Error = "Invalid Email Address"
		}
	}
	Pages.Write(w, r, "register.tmpl", &form)
}

func (u *user) Login(w http.ResponseWriter, r *http.Request) {
	if Session.GetLogin(r) > 0 {
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	var form struct {
		Email, Error string
	}
	var ok bool
	r.ParseForm()
	if _, ok = r.Form["email"]; ok {
		form.Email = r.Form.Get("email")
		uid, err := Users.UserID(form.Email)
		if err != nil {
			//?
		} else {
			if Users.LoginUser(uid, r.Form.Get("password")) == nil {
				Session.SetLogin(w, uid)
				http.Redirect(w, r, "/user/", http.StatusFound)
				return
			}
		}
		form.Error = "Unknown Email Address or Invalid Password"
	}
	Pages.Write(w, r, "login.tmpl", &form)
}

func (u *user) Logout(w http.ResponseWriter, r *http.Request) {
	Session.ClearLogin(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
