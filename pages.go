package main

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/MJKWoolnough/errors"
	vpages "github.com/MJKWoolnough/pages"
)

var Pages pages

const OutputTemplate = "Output"

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
	return nil
}

func (p *pages) RegisterTemplate(path string) error {
	return p.Pages.RegisterFile(path, filepath.Join(*filesDir, path))
}

func (p *pages) Write(w http.ResponseWriter, r *http.Request, templateName string, body interface{}) {
	w.Header().Set("Content-Type", "text/html")
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		userID = Session.GetLogin(r)
	}
	basket, ok := r.Context().Value("basket").(*Basket)
	if !ok {
		basket = Session.LoadBasket(r)
	}
	if err := p.Pages.Write(w, r, templateName, struct {
		LoggedIn bool
		*Basket
		Body interface{}
	}{
		LoggedIn: userID > 0,
		Basket:   basket,
		Body:     body,
	}); err != nil {
		logger.Printf("error writing template: %s", err)
	}
}

func NewPageBytes(title, style string, body template.HTML) *vpages.Bytes {
	return Pages.Bytes(title, style, body)
}

func NewPageFile(title, style, filename string) *vpages.File {
	return Pages.File(title, style, filename)
}
