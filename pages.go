package main

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/MJKWoolnough/errors"
	vpages "github.com/MJKWoolnough/pages"
)

var Pages pages

type pages struct {
	*vpages.Pages
}

func (p *pages) Init() error {
	var err error
	p.Pages, err = vpages.New(filepath.Join(*filesDir, "template.tmpl"))
	if err != nil {
		return errors.WithContext("error initialising pages: ", err)
	}
	p.Pages.StaticString(vpages.StaticTemplate)
	p.Pages.Hook(vpages.HookFn(writeHook))
	return nil
}

func (p *pages) RegisterTemplate(path string) error {
	return p.Pages.RegisterFile(path, filepath.Join(*filesDir, path))
}

type pageData struct {
	LoggedIn bool
	*Basket
	Body interface{}
}

func writeHook(w http.ResponseWriter, r *http.Request, body interface{}) interface{} {
	w.Header().Set("Content-Type", "text/html")
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		userID = Session.GetLogin(r)
	}
	basket, ok := r.Context().Value("basket").(*Basket)
	if !ok {
		basket = Session.LoadBasket(r)
	}
	return pageData{
		LoggedIn: userID > 0,
		Basket:   basket,
		Body:     body,
	}
}

func NewPageBytes(title, style string, body template.HTML) *vpages.Bytes {
	return Pages.Bytes(title, style, body)
}

func NewPageFile(title, style, filename string) *vpages.File {
	return Pages.File(title, style, filepath.Join(*filesDir, filename))
}
