package main

import (
	"context"
	"net/http"
	"path"
	"strings"
)

var Admin admin

type admin struct {
	configT string
}

func (a *admin) init() {
	a.configT = path.Join(*filesDir, "admin", "config.tmpl")
	Pages.RegisterTemplate(a.configT)
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
	switch strings.TrimPrefix(r.URL.Path, "/admin/") {
	case "", "index.html":
		a.index(w, r)
	case "config.html":
		a.config(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (a *admin) index(w http.ResponseWriter, r *http.Request) {
	Pages.Write(w, r,
		PageHeader{
			Title: "CARMEssential - Admin",
			Style: "admin",
		},
		Body{
			Template: OutputTemplate,
			Data:     "ADMIN INDEX",
		},
	)
}

func (a *admin) config(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if _, ok := r.Form["delete"]; ok {
		Config.Remove(r.Form.Get("delete"))
	} else {
		for param := range r.Form {
			if strings.HasPrefix(param, "k_") {
				Config.Set(r.Form.Get(param), r.Form.Get("v_"+param[2:]))
			}
		}
	}
	Pages.Write(w, r,
		PageHeader{
			Title: "CARMEssential - Admin - Config",
			Style: "admin",
		},
		Body{
			Template: a.configT,
			Data:     Config.AsSlice(),
		},
	)
}
