package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strings"
)

var Admin admin

type admin struct {
	configT *template.Template
}

func (a *admin) init() {
	a.configT = template.Must(template.ParseFiles(path.Join(*filesDir, "admin", "config.tmpl")))
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uid := Session.GetLogin(r)
	r = r.WithContext(context.WithValue(r.Context(), "userID", uid))
	if !Users.IsAdmin(uid) {
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
	r.ParseForm()
	if del, ok := r.Form["delete"]; ok {
		Config.Remove(del)
	} else {
		for param := range r.Form {
			if strings.HasPrefix(param, "k_") {
				Config.Set(r.Form.Get(param), r.Form.Get("v_"+param[2:]))
			}
		}
	}
	a.configT.Execute(w, Config.AsSlice())
}
