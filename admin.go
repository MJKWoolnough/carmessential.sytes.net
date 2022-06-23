package main

import "net/http"

type admin struct{}

func (admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func init() {
	http.Handle("/admin", &admin{})
}
