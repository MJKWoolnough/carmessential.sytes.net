package main

import (
	"net/http"
	"os"
)

type admin struct {
	username, password string
}

func (admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func init() {
	user := os.Getenv("adminUser")
	pass := os.Getenv("adminPass")
	if user != "" && pass != "" {
		http.Handle("/admin", &admin{
			username: user,
			password: pass,
		})
	}
}
