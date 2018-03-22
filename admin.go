package main

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/MJKWoolnough/errors"
)

var Admin admin

type admin struct {
	configT, categoriesT, editCategoryT, treatmentsT string
}

func (a *admin) Init() error {
	for _, tmpl := range [...]struct {
		template *string
		path     string
	}{
		{&a.configT, filepath.Join("admin", "config.tmpl")},
		{&a.categoriesT, filepath.Join("admin", "categories.tmpl")},
		{&a.editCategoryT, filepath.Join("admin", "editCategory.tmpl")},
		{&a.treatmentsT, filepath.Join("admin", "treatments.tmpl")},
	} {
		*tmpl.template = tmpl.path
		if err := Pages.RegisterTemplate(tmpl.path); err != nil {
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
	Pages.Write(w, r, "", "ADMIN INDEX")
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
		Pages.Write(w, r, a.configT, configSlice)
	}
}

func (a *admin) categories(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if _, ok := r.PostForm["set"]; ok {
		idStr := r.PostForm.Get("id")
		var (
			id  uint64
			err error
		)
		if idStr != "" {
			id, err = strconv.ParseUint(idStr, 10, 64)
		}
		if err == nil {
			category, exists := Treatments.GetCategory(uint(id))
			if exists || id == 0 {
				Pages.Write(w, r, a.editCategoryT,
					struct {
						Category
						NameError, OrderError string
					}{
						Category: category,
					},
				)
				return
			}
		}
	}
	Pages.Write(w, r, a.categoriesT, Treatments.categories)
}

func (a *admin) treatments(w http.ResponseWriter, r *http.Request) {

}
