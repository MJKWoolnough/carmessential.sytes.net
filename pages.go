package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var Pages pages

type pages struct {
	headerA, headerB, headerC, loggedIn, loggedOut, preBasket, noBasket, postBasket, footer []byte
}

func (p *pages) init(headerA, headerB, headerC, loggedIn, loggedOut, preBasket, noBasket, postBasket, footer string) error {
	var err error
	for filename, data := range map[string]*[]byte{
		headerA:    &p.headerA,
		headerB:    &p.headerB,
		headerC:    &p.headerC,
		loggedIn:   &p.loggedIn,
		loggedOut:  &p.loggedOut,
		preBasket:  &p.preBasket,
		noBasket:   &p.noBasket,
		postBasket: &p.postBasket,
		footer:     &p.footer,
	} {
		if *data, err = loadFile(filename); err != nil {
			return err
		}
	}
	return nil
}

func loadFile(filename string) ([]byte, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, fi.Size())
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	return buf, f.Close()
}

func (p *pages) WriteHeader(w http.ResponseWriter, r *http.Request, ph PageHeader) {
	w.Header().Set("Content-Type", "text/html")
	userID, ok := r.Context().Value("userID").(uint64)
	if !ok {
		userID = Session.GetLogin(r)
	}
	basket, ok := r.Context().Value("basket").(*Basket)
	if !ok {
		basket = Session.LoadBasket(r)
	}
	w.Write(p.headerA)
	w.Write(ph.Title)
	w.Write(p.headerB)
	w.Write(ph.Style)
	w.Write(p.headerC)
	if userID == 0 {
		w.Write(p.loggedOut)
	} else {
		w.Write(p.loggedIn)
	}
	w.Write(p.preBasket)
	if ph.WriteBasket && !basket.IsEmpty() {
		basket.WriteTo(w)
	} else {
		w.Write(p.noBasket)
	}
	w.Write(p.postBasket)
}

func (p *pages) WriteFooter(w http.ResponseWriter) {
	w.Write(p.footer)
}

type PageHeader struct {
	Title, Style []byte
	WriteBasket  bool
}

type PageBytes struct {
	PageHeader
	Page []byte
}

func NewPageBytes(title, style string, data []byte, showBasket bool) *PageBytes {
	return &PageBytes{
		PageHeader: PageHeader{
			Title:       []byte(title),
			Style:       []byte(style),
			WriteBasket: showBasket,
		},
		Page: data,
	}
}

func (p *PageBytes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	Pages.WriteHeader(w, r, p.PageHeader)
	w.Write(p.Page)
	Pages.WriteFooter(w)
}

type PageFile struct {
	PageHeader
	Page string
}

func NewPageFile(title, style, filename string, showBasket bool) *PageFile {
	return &PageFile{
		PageHeader: PageHeader{
			Title:       []byte(title),
			Style:       []byte(style),
			WriteBasket: showBasket,
		},
		Page: filename,
	}
}

func (p *PageFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(p.Page)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	Pages.WriteHeader(w, r, p.PageHeader)
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		io.Copy(w, f)
		f.Close()
	}
	Pages.WriteFooter(w)
}
