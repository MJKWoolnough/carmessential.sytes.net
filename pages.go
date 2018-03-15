package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MJKWoolnough/errors"
)

var Pages pages

const OutputTemplate = "Output"

type pages struct {
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

func (p *pages) init(templateFile string) error {
	bufStr, err := loadFile(templateFile)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error loading main template file (%q): ", templateFile), err)
	}
	splitStr := strings.SplitN(bufStr, "{{/* TEMPLATES HERE */}}", 2)
	if len(splitStr) != 2 {
		return errors.Error("invalid template")
	}
	p.templateT = template.New("")
	_, err = p.templateT.New(OutputTemplate).Parse("{{.}}")
	if err != nil {
		return errors.WithContext("error creating Output template: ", err)
	}
	p.templateData[0] = splitStr[0]
	p.templateData[1] = splitStr[1]
	return p.parseMain()
}

var mainTemplateBuilder strings.Builder

func (p *pages) parseMain() error {
	mainTemplateBuilder.Reset()
	mainTemplateBuilder.WriteString(p.templateData[0])
	mainTemplateBuilder.WriteString("{{if eq .Body.Template \"")
	mainTemplateBuilder.WriteString(OutputTemplate)
	mainTemplateBuilder.WriteString("\"}}{{template \"")
	mainTemplateBuilder.WriteString(OutputTemplate)
	mainTemplateBuilder.WriteString("\" .Body.Data}}")
	for _, tmpl := range p.templates {
		mainTemplateBuilder.WriteString("{{else if eq .Body.Template \"")
		mainTemplateBuilder.WriteString(tmpl)
		mainTemplateBuilder.WriteString("\"}}{{template \"")
		mainTemplateBuilder.WriteString(tmpl)
		mainTemplateBuilder.WriteString("\" .Body.Data}}")
	}
	mainTemplateBuilder.WriteString("{{end}}")
	mainTemplateBuilder.WriteString(p.templateData[1])
	if _, err := p.templateT.Parse(mainTemplateBuilder.String()); err != nil {
		return errors.WithContext("error building main template: ", err)
	}
	return nil
}

func (p *pages) RegisterTemplate(filename string) error {
	data, err := loadFile(filename)
	if err != nil {
		return errors.WithContext(fmt.Sprintf("error loading template file (%q): ", filename), err)
	}
	if _, err = p.templateT.New(filename).Parse(data); err != nil {
		return errors.WithContext(fmt.Sprintf("error parsing template %q: ", filename), err)
	}
	p.templates = append(p.templates, filename)
	return p.parseMain()
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
