package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/MJKWoolnough/httpwrap"
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
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	return buf, f.Close()
}

type wrappedPage struct {
	http.Handler
	writeBasket bool
}

func (p *Pages) Wrap(h http.Handler) http.Handler {
	return wrappedPage{
		Handler:     h,
		printBasket: true,
	}
}

func (p *Pages) SemiWrap(h http.Handler) http.Handler {
	return wrappedPage{
		Handler:     h,
		writeBasket: false,
	}
}

func (wp wrappedPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := Session.GetLogin(r)
	basket := Session.LoadBasket(r)
	wBasket := basket
	if !wp.writeBasket {
		wBasket = nil
	}
	ww := wrappedWriter{
		ResponseWriter: w,
		basket:         wBasket,
		loggedIn:       userID > 0,
	}
	wp.Handler.ServeHTTP(
		httpwrap.Wrap(
			w,
			httpwrap.OverrideWriter(ww),
		),
		r.WithContext(
			context.WithValue(
				context.WithValue(
					r.Context(),
					"basket", basket,
				),
				"userID", userID,
			),
		),
	)
	if ww.written {
		w.Write(Pages.footer)
	}
}

type wrappedWriter struct {
	http.ResponseWriter
	basket   *Basket
	loggedIn bool
	written  bool
}

func (w *wrappedWriter) Write(p []byte) (int, error) {
	if !w.written {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "text/html")
		}
		w.ResponseWriter.Write(Pages.header)
		if w.loggedIn {
			w.ResponseWriter.Write(Pages.loggedIn)
		} else {
			w.ResponseWriter.Write(Pages.loggedOut)
		}
		if p.Basket.IsEmpty() {
			w.ResponseWriter.Write(Pages.noBasket)
		} else {
			w.Basket.WriteTo(w.ResponseWriter)
		}
		w.ResponseWriter.Write(Pages.postBasket)
		w.written = true
	}
	return w.ResponseWriter.Write(p)
}
