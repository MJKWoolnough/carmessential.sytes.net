package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
	"vimagination.zapto.org/form"
	"vimagination.zapto.org/jsonrpc"
	"vimagination.zapto.org/sessions"

	_ "github.com/mattn/go-sqlite3"
)

var (
	//go:embed admin.html
	adminPage []byte
	//go:embed login.html
	loginPage     string
	adminOnline   uint32
	oneAdmin      = []byte("{\"id\":-1,\"error\":{\"code\":1,\"message\":\"admin online\"}}")
	goodAdmin     = []byte("{\"id\":-1,\"result\":0}")
	loginTemplate *template.Template
	db            *sql.DB

	hf     sync.RWMutex
	header string
	footer string
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
		pass := fmt.Sprintf("%x", sha256.Sum256([]byte(l.Password)))
		if l.Username == a.username && pass == a.password {
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
		w.Write(adminPage)
	} else {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, header)
		loginTemplate.Execute(w, l)
		io.WriteString(w, footer)
	}
}

func (a *admin) serveConn(wconn *websocket.Conn) {
	hf.RLock()
	fmt.Fprintf(wconn, "{\"id\":-2,\"result\":{\"header\":%q,\"footer\":%q}}", header, footer)
	hf.RUnlock()
	if atomic.CompareAndSwapUint32(&adminOnline, 0, 1) {
		wconn.Write(goodAdmin)
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
	if a, err := adminInit(); err == nil {
		http.Handle("/admin", a)
	}
}

func adminInit() (*admin, error) {
	user := os.Getenv("adminUser")
	if user == "" {
		return nil, errors.New("no admin username")
	}
	pass := os.Getenv("adminPass")
	if len(pass) != 64 {
		return nil, errors.New("no admin password")
	}
	adminDB := os.Getenv("adminDB")
	if adminDB == "" {
		return nil, errors.New("no admin database")
	}
	key, _ := base64.StdEncoding.DecodeString(os.Getenv("adminKey"))
	if len(key) != 16 {
		return nil, errors.New("no admin key")
	}
	data, _ := base64.StdEncoding.DecodeString(os.Getenv("adminData"))
	if len(data) != 32 {
		return nil, errors.New("no admin data")
	}
	store, err := sessions.NewCookieStore(key, sessions.HTTPOnly(), sessions.Path("/"), sessions.Name("admin"), sessions.Expiry(time.Hour*24*30))
	if err != nil {
		return nil, err
	}
	db, err = sql.Open("sqlite3", adminDB)
	if err != nil {
		return nil, err
	}
	a := &admin{
		username:    user,
		password:    pass,
		CookieStore: store,
		sessionData: data,
	}
	a.rpc = websocket.Handler(a.serveConn)
	loginTemplate, _ = template.New("login").Parse(loginPage)
	for _, ct := range []string{
		"[Settings]([Version] INTEGER DEFAULT 0, [Header] TEXT NOT NULL DEFAULT '', [Footer] TEXT NOT NULL DEFAULT '');",
	} {
		if _, err = db.Exec("CREATE TABLE IF NOT EXISTS " + ct); err != nil {
			return nil, err
		}
	}
	count := 0
	db.QueryRow("SELECT COUNT(1) FROM [Settings];").Scan(&count)
	if count == 0 {
		if _, err = db.Exec("INSERT INTO [Settings] ([Version]) VALUES (0);"); err != nil {
			return nil, err
		}
	} else {
		if err = db.QueryRow("SELECT [Header], [Footer] [Settings];").Scan(&header, &footer); err != nil {
			return nil, err
		}
	}
	return a, nil
}
