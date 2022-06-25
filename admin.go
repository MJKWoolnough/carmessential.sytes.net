package main

import (
	"bytes"
	"database/sql"
	_ "embed"
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

	_ "github.com/mattn/go-sqlite3"
)

var (
	//go:embed admin.html
	adminPage []byte
	//go:embed login.html
	loginPage     string
	adminOnline   uint32
	oneAdmin      = []byte("{\"id\":-1,\"error\":{\"code\":1,\"message\":\"admin online\"}}")
	goodAdmin     = []byte("{\"id\":-1,\"result\": 0}")
	loginTemplate *template.Template
	db            *sql.DB
	header        string
	footer        string
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
		w.Write(adminPage)
	} else {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, header)
		loginTemplate.Execute(w, l)
		io.WriteString(w, footer)
	}
}

func (a *admin) serveConn(wconn *websocket.Conn) {
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
	user := os.Getenv("adminUser")
	pass := os.Getenv("adminPass")
	adminDB := os.Getenv("adminDB")
	key, _ := base64.StdEncoding.DecodeString(os.Getenv("adminKey"))
	data, _ := base64.StdEncoding.DecodeString(os.Getenv("adminData"))
	if user != "" && pass != "" && len(key) == 16 && len(data) == 32 {
		store, err := sessions.NewCookieStore(key, sessions.HTTPOnly(), sessions.Path("/"), sessions.Name("admin"), sessions.Expiry(time.Hour*24*30))
		if err == nil {
			db, err = sql.Open("sqlite3", adminDB)
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
				for _, ct := range []string{
					"[Settings]([Version] INTEGER DEFAULT 0, [Header] TEXT NOT NULL DEFAULT '', [Footer] TEXT NOT NULL DEFAULT '');",
				} {
					db.Exec("CREATE TABLE IF NOT EXISTS " + ct)
				}
				count := 0
				db.QueryRow("SELECT COUNT(1) FROM [Settings];").Scan(&count)
				if count == 0 {
					db.Exec("INSERT INTO [Settings] ([Version]) VALUES (0);")
				} else {
					db.QueryRow("SELECT [Header], [Footer] [Settings];").Scan(&header, &footer)
				}
			}
		}
	}
}
