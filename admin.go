package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
	"vimagination.zapto.org/form"
	"vimagination.zapto.org/jsonrpc"
	"vimagination.zapto.org/sessions"
)

var (
	adminOnline   uint32
	oneAdmin      = []byte("{\"id\":-1,\"error\":\"admin online\"}")
	loginTemplate *template.Template
)

type login struct {
	Username string `form:"username,post"`
	Password string `form:"password,post"`
	Error    string `form:"-"`
}

type admin struct {
	username, password string
	*sessions.CookieStore
	sessionData []byte
	rpc         websocket.Handler
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isAdmin := bytes.Equal(a.CookieStore.Get(r), a.sessionData)
	var l login
	if !isAdmin && r.Method == http.MethodPost {
		form.Process(r, &l)
		if l.Username == a.username && l.Password == a.password {
			a.CookieStore.Set(w, a.sessionData)
			isAdmin = true
		} else {
			l.Error = "Invalid Username or Password"
		}
	}
	if isAdmin {
		if r.Header.Get("Upgrade") == "websocket" {
			a.rpc.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, adminPage)
	} else {
		w.Header().Set("Content-Type", "text/html")
		loginTemplate.Execute(w, l)
	}
}

func (a *admin) serveConn(wconn *websocket.Conn) {
	if atomic.CompareAndSwapUint32(&adminOnline, 0, 1) {
		jsonrpc.New(wconn, a).Handle()
		atomic.StoreUint32(&adminOnline, 0)
	} else {
		wconn.Write(oneAdmin)
	}
}

func (a *admin) HandleRPC(method string, data json.RawMessage) (interface{}, error) {
	return nil, nil
}

func init() {
	user := os.Getenv("adminUser")
	pass := os.Getenv("adminPass")
	key, _ := base64.StdEncoding.DecodeString(os.Getenv("adminKey"))
	data, _ := base64.StdEncoding.DecodeString(os.Getenv("adminData"))
	if user != "" && pass != "" && len(key) == 16 && len(data) == 32 {
		store, err := sessions.NewCookieStore(key, sessions.HTTPOnly(), sessions.Path("/"), sessions.Name("admin"), sessions.Expiry(time.Hour*24*30))
		if err == nil {
			a := &admin{
				username:    user,
				password:    pass,
				CookieStore: store,
				sessionData: data,
			}
			a.rpc = websocket.Handler(a.serveConn)
			http.Handle("/admin", a)
			loginTemplate, _ = template.New("login").Parse(loginPage)
		}
	}
}
