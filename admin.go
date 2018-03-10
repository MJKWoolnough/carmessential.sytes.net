package main

import (
	"fmt"
	"net/http"
	"strings"
)

var Admin admin

type admin struct {
}

func (a *admin) init() {

}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if uid := Session.GetLogin(r); !Session.IsAdmin(uid) {
		if uid == 0 {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	Pages.WriteHeader(w, r, PageHeader{
		Title:       []byte("CARMEssential - Admin"),
		Style:       []byte("admin"),
		WriteBasket: false,
	})
	switch strings.TrimPrefix(r.URL.Path, "/admin/") {
	case "", "index.html":
		a.index(w, r)
	case "config.html":
		a.config(w, r)
	default:
		fmt.Fprintln(w, "404")
	}
	Pages.WriteFooter(w)
}

func (a *admin) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "INDEX")
}

func (a *admin) config(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "CONFIG")
}
