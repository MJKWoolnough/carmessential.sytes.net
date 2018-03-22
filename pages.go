package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MJKWoolnough/errors"
)

var Pages pages

const OutputTemplate = "Output"

type pages struct {
	mu              sync.RWMutex
	defaultTemplate *template.Template
	templates       map[string]*template.Template
}

func (p *pages) registerTemplate(name, filename string) error {
	templateSrc, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error loading template (%q): ", filename), err)
	}
	dtc, err := p.defaultTemplate.Clone()
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error cloning template (%q): ", filename), err)
	}
	p.templates[name], err = dtc.Parse(string(templateSrc))
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error initialising template (%q): ", filename), err)
	}
	return nil
}

func (p *pages) Init() error {
	templateSrc, err := ioutil.ReadFile(filepath.Join(*filesDir, "template.tmpl"))
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error loading template (%q): ", "template.tmpl"), err)
	}
	p.defaultTemplate, err = template.New("").Parse(string(templateSrc))
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error initialising template (%q): ", "template.tmpl"), err)
	}
	p.templates = map[string]*template.Template{"": p.defaultTemplate}
	dtc, _ := p.defaultTemplate.Clone()
	dtc.Parse("{{define \"title\"}}{{.Title}}{{end}}{{define \"style\"}}{{.Style}}{{end}}{{define \"body\"}}\n{{.Body}}{{end}}")
	p.templates["dynamic"] = dtc
	return nil
}

func (p *pages) RegisterTemplate(filename string) error {
	return p.registerTemplate(filename, filepath.Join(*filesDir, filename))
}

func (p *pages) Rebuild() error {
	p.mu.Lock()
	oldTemplate := p.defaultTemplate
	oldTemplates := p.templates
	if err := p.Init(); err != nil {
		p.defaultTemplate = oldTemplate
		p.templates = oldTemplates
		p.mu.Unlock()
		return errors.WithContext("error reloading templates: ", err)
	}
	for filename := range oldTemplates {
		switch filename {
		case "", "dynamic":
		default:
			if err := p.RegisterTemplate(filename); err != nil {
				p.defaultTemplate = oldTemplate
				p.templates = oldTemplates
				p.mu.Unlock()
				return errors.WithContext("error reloading templates: ", err)
			}
		}
	}
	p.mu.Unlock()
	return nil
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
	p.mu.RLock()
	tmpl, ok := p.templates[templateName]
	if !ok {
		tmpl = p.defaultTemplate
	}
	if err := tmpl.Execute(w, struct {
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
	p.mu.RUnlock()
}

type pageData struct {
	Title, Style string
	Body         template.HTML
}

type PageBytes struct {
	pageData pageData
}

func NewPageBytes(title, style string, body template.HTML) *PageBytes {
	return &PageBytes{
		pageData: pageData{
			Title: title,
			Style: style,
			Body:  body,
		},
	}
}

func (p *PageBytes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Pages.Write(w, r, "dynamic", p.pageData)
}

type PageFile struct {
	mu           sync.RWMutex
	pageData     pageData
	Filename     string
	LastModified time.Time
}

func NewPageFile(title, style, filename string) *PageFile {
	return &PageFile{
		pageData: pageData{
			Title: title,
			Style: style,
		},
		Filename: filepath.Join(*filesDir, filename),
	}
}

func (p *PageFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats, err := os.Stat(p.Filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if modtime := stats.ModTime(); modtime.After(p.LastModified) {
		p.mu.Lock()
		stats, err = os.Stat(p.Filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if modtime = stats.ModTime(); modtime.After(p.LastModified) { // in case another goroutine has changed it already
			data, err := ioutil.ReadFile(p.Filename)
			if err != nil {
				p.mu.Unlock()
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				p.pageData.Body = template.HTML(data)
				p.LastModified = modtime
			}
		}
		p.mu.Unlock()
	}
	p.mu.RLock()
	body := p.pageData
	p.mu.RUnlock()
	Pages.Write(w, r, "dynamic", body)
}
