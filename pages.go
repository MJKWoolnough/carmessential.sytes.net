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
	if p.header, err = loadFile(header); err != nil {
		return err
	}
	if p.header, err = loadFile(loggedIn); err != nil {
		return err
	}
	if p.header, err = loadFile(loggedOut); err != nil {
		return err
	}
	if p.header, err = loadFile(preBasket); err != nil {
		return err
	}
	if p.header, err = loadFile(noBasket); err != nil {
		return err
	}
	if p.header, err = loadFile(postBasket); err != nil {
		return err
	}
	if p.header, err = loadFile(footer); err != nil {
		return err
	}
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
	var basket []byte // process basket

	p.ResponseWriter.Write(Pages.header)
	if p.UserID != 0 {
		p.ResponseWriter.Write(Pages.loggedIn)
	} else {
		p.ResponseWriter.Write(Pages.loggedOut)
	}
	p.ResponseWriter.Write(Pages.preBasket)
	p.ResponseWriter.Write(basket)
	p.ResponseWriter.Write(Pages.postBasket)
	p.ResponseWriter.Write(p.Body)
	p.ResponseWriter.Write(Pages.footer)

}
