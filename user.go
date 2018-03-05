package main

import (
	"encoding/base64"
	"html/template"
	"net/http"
	"time"

	"github.com/MJKWoolnough/authenticate"
	"github.com/MJKWoolnough/memio"
)

var User user

type user struct {
	loginT, registerT, emailT *template.Template
	from                      string
	registerCodec             *authenticate.Codec
}

func (u *user) init(login, register, email, from, registerKey string) error {
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
	u.registerCodec, err = authenticate.NewCodec([]byte(registerKey), time.Hour*24)
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
