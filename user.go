package main

import (
	"html/template"
	"net/http"
)

var User user

type user struct {
	Login, Register *template.Template
}

func (u *user) init(login, register string) error {
	loginTemplate, err := template.ParseFiles(login)
	registerTemplate, err := template.ParseFiles(register)
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
		Title:       "CARMEssential - User Area",
		Style:       "user",
		WriteBasket: true,
	})
	w.Write([]byte("USER"))
	Pages.WriteFooter(w)
}

func (u *user) Register(w http.ResponseWriter, r *http.Request) {
	if Session.GetLogin(r) > 0 {
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	Pages.WriteHeader(w, r, PageHeader{
		Title:       "CARMEssential - Register",
		Style:       "user",
		WriteBasket: true,
	})
	w.Write([]byte("REGISTER"))
	Pages.WriteFooter(w)
}

func (u *user) Login(w http.ResponseWriter, r *http.Request) {
	if Session.GetLogin(r) > 0 {
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	Pages.WriteHeader(w, r, PageHeader{
		Title:       "CARMEssential - Login",
		Style:       "user",
		WriteBasket: true,
	})
	w.Write([]byte("LOGIN"))
	Pages.WriteFooter(w)
}

func (u *user) Logout(w http.ResponseWriter, r *http.Request) {
	Session.ClearLogin(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
