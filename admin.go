package main

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"os"
	"time"

	"vimagination.zapto.org/form"
	"vimagination.zapto.org/sessions"
)

type login struct {
	Username string `form:"username,post"`
	Password string `form:"password,post"`
}

type admin struct {
	username, password string
	*sessions.CookieStore
	sessionData []byte
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isAdmin := bytes.Equal(a.CookieStore.Get(r), a.sessionData)
	if !isAdmin {
		l := new(login)
		form.Process(r, l)
		if l.Username == a.username && l.Password == a.password {
			a.CookieStore.Set(w, a.sessionData)
			isAdmin = true
		}
	}
	if isAdmin {
	} else {
	}
}

func init() {
	user := os.Getenv("adminUser")
	pass := os.Getenv("adminPass")
	key, _ := base64.StdEncoding.DecodeString(os.Getenv("adminKey"))
	data, _ := base64.StdEncoding.DecodeString(os.Getenv("adminData"))
	if user != "" && pass != "" && len(key) == 16 && len(data) != 32 {
		store, err := sessions.NewCookieStore(key, sessions.HTTPOnly(), sessions.Path("/"), sessions.Name("admin"), sessions.Expiry(time.Hour*24*30))
		if err == nil {
			http.Handle("/admin", &admin{
				username:    user,
				password:    pass,
				CookieStore: store,
				sessionData: data,
			})
		}
	}
}
