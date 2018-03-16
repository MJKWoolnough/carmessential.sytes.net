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

func (a *admin) init() error {
	a.configT = path.Join(*filesDir, "admin", "config.tmpl")
	return Pages.RegisterTemplate(a.configT)
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
