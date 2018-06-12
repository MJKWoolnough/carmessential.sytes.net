package main

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"vimagination.zapto.org/errors"
)

var Admin admin

type admin struct {
	indexT, configT, categoriesT, editCategoryT, treatmentsT, editTreatmentT, templatesT string
}

func (a *admin) Init() error {
	for _, tmpl := range [...]struct {
		template *string
		path     string
	}{
		{&a.indexT, filepath.Join("admin", "index.tmpl")},
		{&a.configT, filepath.Join("admin", "config.tmpl")},
		{&a.categoriesT, filepath.Join("admin", "categories.tmpl")},
		{&a.editCategoryT, filepath.Join("admin", "editCategory.tmpl")},
		{&a.treatmentsT, filepath.Join("admin", "treatments.tmpl")},
		{&a.editTreatmentT, filepath.Join("admin", "editTreatment.tmpl")},
		{&a.templatesT, filepath.Join("admin", "templates.tmpl")},
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
	case "templates.html":
		a.templates(w, r)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (a *admin) index(w http.ResponseWriter, r *http.Request) {
	Pages.Write(w, r, a.indexT, "ADMIN INDEX")
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
			var (
				nameError, orderError string
				category              Category
				exists                bool
			)
			_, nameOK := r.PostForm["name"]
			_, orderOK := r.PostForm["order"]
			if nameOK && orderOK {
				exists = true
				category.ID = uint(id)
				category.Name = r.PostForm.Get("name")
				if Treatments.GetCategoryID(category.Name) != category.ID {
					nameError = "category already exists"
				}
				order, err := strconv.ParseUint(r.PostForm.Get("order"), 10, 64)
				if err != nil {
					orderError = err.Error()
				}
				category.Order = uint(order)
				category.AdminOnly, _ = strconv.ParseBool(r.PostForm.Get("adminOnly"))
				if nameError == "" && orderError == "" {
					Treatments.SetCategory(&category)
					http.Redirect(w, r, "/admin/categories.html", http.StatusFound)
					return
				}
			} else {
				category, exists = Treatments.GetCategory(uint(id))
			}
			if exists || id == 0 {
				Pages.Write(w, r, a.editCategoryT,
					struct {
						Category
						NameError, OrderError string
					}{
						Category:   category,
						NameError:  nameError,
						OrderError: orderError,
					},
				)
				return
			}
		}
	} else if _, ok := r.PostForm["delete"]; ok {
		id, err := strconv.ParseUint(r.PostForm.Get("id"), 10, 64)
		if err == nil {
			Treatments.RemoveCategory(uint(id))
			http.Redirect(w, r, "/admin/categories.html", http.StatusFound)
			return
		}
	}
	Pages.Write(w, r, a.categoriesT, Treatments.GetCategories())
}

func (a *admin) treatments(w http.ResponseWriter, r *http.Request) {
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
			var (
				nameError, orderError, categoryError, priceError, timeError, descriptionError string
				treatment                                                                     Treatment
				exists                                                                        bool
			)
			_, nameOK := r.PostForm["name"]
			_, orderOK := r.PostForm["order"]
			_, categoryOK := r.PostForm["catID"]
			_, priceOK := r.PostForm["price"]
			_, timeOK := r.PostForm["time"]
			_, descriptionOK := r.PostForm["description"]
			if nameOK && orderOK && categoryOK && priceOK && timeOK && descriptionOK {
				exists = true
				treatment.ID = uint(id)
				treatment.Name = r.PostForm.Get("name")
				if Treatments.GetTreatmentID(treatment.Name) != treatment.ID {
					nameError = "treatment already exists"
				}
				order, err := strconv.ParseUint(r.PostForm.Get("order"), 10, 64)
				if err != nil {
					orderError = err.Error()
				}
				treatment.Order = uint(order)
				category, err := strconv.ParseUint(r.PostForm.Get("catID"), 10, 64)
				if err != nil {
					categoryError = err.Error()
				} else if _, ok := Treatments.GetCategory(uint(category)); !ok {
					categoryError = "unknown category"
				} else {
					treatment.Category = uint(category)
				}
				if price, err := strconv.ParseFloat(r.PostForm.Get("price"), 32); err != nil {
					priceError = err.Error()
				} else {
					treatment.Price = uint(price * 100)
				}
				if duration, err := strconv.ParseUint(r.PostForm.Get("time"), 10, 64); err != nil {
					timeError = err.Error()
				} else if duration%15 != 0 {
					timeError = "duration needs to be a multiple of 15 minutes"
				} else {
					treatment.Duration = time.Duration(duration) * time.Minute
				}
				treatment.DescriptionSrc = r.PostForm.Get("description")
				if treatment.DescriptionSrc == "" {
					descriptionError = "need a description"
				}
				if nameError == "" && orderError == "" && categoryError == "" && priceError == "" && timeError == "" && descriptionError == "" {
					Treatments.SetTreatment(&treatment)
					http.Redirect(w, r, "/admin/treatments.html", http.StatusFound)
					return
				}
			} else {
				treatment, exists = Treatments.GetTreatment(uint(id))
			}
			if exists || id == 0 {
				Pages.Write(w, r, a.editTreatmentT,
					struct {
						Treatment
						Categories                                                                    []Category
						Duration                                                                      uint
						Price                                                                         float32
						NameError, OrderError, CategoryError, PriceError, TimeError, DescriptionError string
					}{
						Treatment:        treatment,
						Categories:       Treatments.GetCategories(),
						Duration:         uint(treatment.Duration / time.Minute),
						Price:            float32(treatment.Price) / 100,
						NameError:        nameError,
						OrderError:       orderError,
						CategoryError:    categoryError,
						PriceError:       priceError,
						TimeError:        timeError,
						DescriptionError: descriptionError,
					},
				)
				return
			}
		}
	} else if _, ok := r.PostForm["delete"]; ok {
		id, err := strconv.ParseUint(r.PostForm.Get("id"), 10, 64)
		if err == nil {
			Treatments.RemoveTreatment(uint(id))
			http.Redirect(w, r, "/admin/treatments.html", http.StatusFound)
			return
		}
	}
	Pages.Write(w, r, a.treatmentsT, Treatments.GetTreatments())
}

func (a *admin) templates(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var data struct {
		Updated bool
		Error   string
	}
	if _, data.Updated = r.PostForm["reload"]; data.Updated {
		if err := Pages.Rebuild(); err != nil {
			data.Error = err.Error()
		}
	}
	Pages.Write(w, r, a.templatesT, data)
}
