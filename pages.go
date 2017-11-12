package main

import "net/http"

var Pages pages

type pages struct {
	header, loggedIn, loggedOut, preBasket, postBasket, footer []byte
}

func (p *pages) init() {

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
