package main

import (
	"context"
	"net/http"
	"path"
	"strings"

	"github.com/MJKWoolnough/errors"
)

var Admin admin

type admin struct {
	configT string
}

func (a *admin) init() error {
	a.configT = path.Join(*filesDir, "admin", "config.tmpl")
	for _, tmpl := range [...]string{
		a.configT,
	} {
		if err := Pages.RegisterTemplate(tmpl); err != nil {
			return errors.WithContext("error registering admin template: ", err)
		}
	}
	return nil
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uid := Session.GetLogin(r)
	if !Users.IsAdmin(uid) {
		if uid == 0 {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/user/", http.StatusFound)
		return
	}
	r = r.WithContext(context.WithValue(r.Context(), "userID", uid))
	switch strings.TrimPrefix(r.URL.Path, "/admin/") {
	case "", "index.html":
		a.index(w, r)
	case "config.html":
		a.config(w, r)
	case "categories.html":
		a.categories(w, r)
	case "treatments.html":
		a.treatments(w, r)
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
	if _, ok := r.PostForm["delete"]; ok {
		Config.Remove(r.PostForm.Get("delete"))
	} else {
		for param := range r.PostForm {
			if strings.HasPrefix(param, "k_") {
				key := r.PostForm.Get(param)
				if key != "" {
					Config.Set(key, r.PostForm.Get("v_"+param[2:]))
				}
			}
		}
	}
	if _, ok := r.PostForm["dynamic"]; !ok {
		configSlice := Config.AsSlice()
		if _, ok := r.PostForm["add"]; ok {
			configSlice = append(configSlice, KeyValue{})
		}
		Pages.Write(w, r,
			PageHeader{
				Title:  "CARMEssential - Admin - Config",
				Style:  "admin",
				Script: "config",
			},
			Body{
				Template: a.configT,
				Data:     configSlice,
			},
		)
	}
}

func (a *admin) categories(w http.ResponseWriter, r *http.Request) {

}

func (a *admin) treatments(w http.ResponseWriter, r *http.Request) {

}
