package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MJKWoolnough/errors"
)

var Pages pages

const OutputTemplate = "Output"

type pages struct {
	mu           sync.RWMutex
	templateT    *template.Template
	templateData [2]string
	templates    []string
}

func loadFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	buf := make([]byte, stat.Size())
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (p *pages) init() error {
	p.templateT = template.New("")
	if err := p.makeOutputTemplate(); err != nil {
		return err
	}
	return p.loadMainTemplate()
}

func (p *pages) makeOutputTemplate() error {
	_, err := p.templateT.New(OutputTemplate).Parse("{{.}}")
	if err != nil {
		return errors.WithContext("error creating Output template: ", err)
	}
	return nil
}

func (p *pages) loadMainTemplate() error {
	bufStr, err := loadFile(filepath.Join(*filesDir, "template.tmpl"))
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error loading main template file (%q): ", p.templateF), err)
	}
	splitStr := strings.SplitN(bufStr, "{{/* TEMPLATES HERE */}}", 2)
	if len(splitStr) != 2 {
		return errors.Error("invalid template")
	}
	p.templateData[0] = splitStr[0]
	p.templateData[1] = splitStr[1]
	return p.buildMain()
}

var mainTemplateBuilder strings.Builder

func (p *pages) buildMain() error {
	mainTemplateBuilder.Reset()
	fmt.Fprintf(&mainTemplateBuilder, "%[1]s{{if eq .Body.Template %[2]q}}{{template %[2]q .Body.Data}}", p.templateData[0], OutputTemplate)
	for _, tmpl := range p.templates {
		fmt.Fprintf(&mainTemplateBuilder, "{{else if eq .Body.Template %[1]q}}{{template %[1]q .Body.Data}}", tmpl)
	}
	fmt.Fprintf(&mainTemplateBuilder, "{{end}}%s", p.templateData[1])
	if _, err := p.templateT.Parse(mainTemplateBuilder.String()); err != nil {
		return errors.WithContext("error building main template: ", err)
	}
	return nil
}

func (p *pages) registerTemplate(filename string) error {
	data, err := loadFile(filename)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error loading template file (%q): ", filename), err)
	}
	if _, err = p.templateT.New(filename).Parse(data); err != nil {
		return errors.WithContext(fmt.Sprintf("error parsing template %q: ", filename), err)
	}
	return nil
}

func (p *pages) RegisterTemplate(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.registerTemplate(filename); err != nil {
		return err
	}
	p.templates = append(p.templates, filename)
	err := p.buildMain()
	return err
}

func (p *pages) Rebuild() error {
	old := p.templateT
	p.mu.Lock()
	defer p.mu.Unlock()
	p.templateT = template.New("")
	if err := p.makeOutputTemplate(); err != nil {
		p.templateT = old
		return err
	}
	for _, tmpl := range p.templates {
		if err := p.registerTemplate(tmpl); err != nil {
			p.templateT = old
			return err
		}
	}
	if err := p.loadMainTemplate(); err != nil {
		p.templateT = old
		return err
	}
	return nil
}

func (p *pages) Write(w http.ResponseWriter, r *http.Request, ph PageHeader, body Body) {
	w.Header().Set("Content-Type", "text/html")
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		userID = Session.GetLogin(r)
	}
	var basket *Basket
	if ph.WriteBasket {
		basket, ok = r.Context().Value("basket").(*Basket)
		if !ok {
			basket = Session.LoadBasket(r)
		}
	}
	p.mu.RLock()
	if err := p.templateT.Execute(w, struct {
		LoggedIn bool
		PageHeader
		*Basket
		Body
	}{
		LoggedIn:   userID > 0,
		PageHeader: ph,
		Basket:     basket,
		Body:       body,
	}); err != nil {
		logger.Printf("error writing template: %s", err)
	}
	p.mu.RUnlock()
}

type Body struct {
	Template string
	Data     interface{}
}

type PageHeader struct {
	Title, Style, Script string
	WriteBasket          bool
}

type PageBytes struct {
	PageHeader
	Body
}

func NewPageBytes(title, style, script string, data template.HTML, showBasket bool) *PageBytes {
	return &PageBytes{
		PageHeader: PageHeader{
			Title:       title,
			Style:       style,
			Script:      script,
			WriteBasket: showBasket,
		},
		Body: Body{
			Template: OutputTemplate,
			Data:     data,
		},
	}
}

func (p *PageBytes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Pages.Write(w, r, p.PageHeader, p.Body)
}

type PageFile struct {
	mu sync.RWMutex
	PageHeader
	Body
	Filename     string
	LastModified time.Time
}

func NewPageFile(title, style, script, filename string, showBasket bool) *PageFile {
	return &PageFile{
		PageHeader: PageHeader{
			Title:       title,
			Style:       style,
			Script:      script,
			WriteBasket: showBasket,
		},
		Body: Body{
			Template: OutputTemplate,
		},
		Filename: filename,
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
			data, err := loadFile(p.Filename)
			if err != nil {
				p.mu.Unlock()
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				p.Data = template.HTML(data)
				p.LastModified = modtime
			}
		}
		p.mu.Unlock()
	}
	p.mu.RLock()
	body := p.Body
	p.mu.RUnlock()
	Pages.Write(w, r, p.PageHeader, body)
}
