package main

import (
	"io"
	"net/http"
	"os"
)

var Pages pages

type pages struct {
	header, loggedIn, loggedOut, preBasket, noBasket, postBasket, footer []byte
}

func (p *pages) init(header, loggedIn, loggedOut, preBasket, noBasket, postBasket, footer string) error {
	var err error
	for filename, data := range map[string]*[]byte{
		header:     &p.header,
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
	buf := make([]byte, 0, fi.Size())
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	_, err := io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	return buf, f.Close()
}

type Page struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	UserID         uint64
	Basket         *Basket
	Body           []byte
}

func (p *pages) Start(w http.ResponseWriter, r *http.Request) Page {
	return Page{
		ResponseWriter: w,
		Request:        r,
		UserID:         session.GetLogin(r),
		Basket:         session.LoadBasket(r),
	}
}

func (p Page) Output() {
	if p.ResponseWriter.Header().Get("Content-Type") != "" {
		p.ResponseWriter.Header().Set("Content-Type", "text/html")
	}
	p.ResponseWriter.Write(Pages.header)
	if p.UserID != 0 {
		p.ResponseWriter.Write(Pages.loggedIn)
	} else {
		p.ResponseWriter.Write(Pages.loggedOut)
	}
	p.ResponseWriter.Write(Pages.preBasket)
	if p.Basket.IsEmpty() {
		p.ResponseWriter.Write(Pages.noBasket)
	} else {
		p.Basket.WriteTo(p.ResponseWriter)
	}
	p.ResponseWriter.Write(Pages.postBasket)
	p.ResponseWriter.Write(p.Body)
	p.ResponseWriter.Write(Pages.footer)
}
