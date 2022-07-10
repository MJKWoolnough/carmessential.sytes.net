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
	"vimagination.zapto.org/memio"
	"vimagination.zapto.org/sessions"

	_ "github.com/mattn/go-sqlite3"
)

type treatment struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Price       uint32 `json:"price"`
	Description string `json:"description"`
	Duration    uint32 `json:"duration"`
}

const (
	setHeaderFooter = iota

	listTreatments
	addTreatment
	setTreatment
	removeTreatment

	bookingsOnDate
	addBooking
	updateBooking
	removeBooking

	totalStmts
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

	statements [totalStmts]*sql.Stmt
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
	switch method {
	case "setHeaderFooter":
		var headfoot [2]string
		if err := json.Unmarshal(data, &headfoot); err != nil {
			return nil, err
		}
		if _, err := statements[setHeaderFooter].Exec(headfoot[0], headfoot[1]); err != nil {
			return nil, err
		}
		hf.Lock()
		header = headfoot[0]
		footer = headfoot[1]
		hf.Unlock()
		generatePages(-1)
		return nil, nil
	case "listTreatments":
		r, err := statements[listTreatments].Query()
		if err != nil {
			return nil, err
		}
		buf := memio.Buffer("[")
		first := true
		for r.Next() {
			var (
				id                       uint64
				price, duration          uint32
				name, group, description string
			)
			if err := r.Scan(id, name, group, price, description, duration); err != nil {
				return nil, err
			}
			if first {
				first = false
			} else {
				buf = append(buf, ',')
			}
			fmt.Fprintf(&buf, "%d,%q,%q,%d,%q,%d", id, name, group, price, description, duration)
		}
		buf = append(buf, ']')
		return json.RawMessage(buf), nil
	case "addTreatment":
		var t treatment
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, err
		}
		r, err := statements[addTreatment].Exec(t.Name, t.Group, t.Price, t.Description, t.Duration)
		if err != nil {
			return nil, err
		}
		id, err := r.LastInsertId()
		if err != nil {
			return nil, err
		}
		generatePages(id)
		return id, nil
	case "setTreatment":
		var t treatment
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, err
		}
		if _, err := statements[setTreatment].Exec(t.Name, t.Group, t.Price, t.Description, t.Duration, t.ID); err != nil {
			return nil, err
		}
		generatePages(int64(t.ID))
		return nil, nil
	case "removeTreatment":
		var id uint32
		if err := json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		if _, err := statements[removeTreatment].Exec(id); err != nil {
			return nil, err
		}
		// remove treatment page
		generatePages(-1)
		return nil, nil
	}
	return nil, errors.New("unknown endpoint")
}

func generatePages(id int64) {
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
		"[Treatments]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Group] TEXT NOT NULL DEFAULT '', [Price] INTEGER NOT NULL, [Description] TEXT NOT NULL DEFAULT '', [Duration] INTEGER NOT NULL);",
		"[Bookings]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Date] INTEGER NOT NULL, [BlockNum] INTEGER NOT NULL, [TotalBlocks] INTEGER NOT NULL, [TreatmentID] INTEGER NOT NULL, [Name] TEXT NOT NULL DEFAULT '', [EmailAddress] NOT NULL DEFAULT '', [PhoneNumber] NOT NULL DEFAULT '', [OrderID] INTEGER NOT NULL);",
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
		if err = db.QueryRow("SELECT [Header], [Footer] FROM [Settings];").Scan(&header, &footer); err != nil {
			return nil, err
		}
	}
	for n, ps := range []string{
		"UPDATE [Settings] SET [Header] = ?, [Footer] = ?;",

		// Treatments
		"SELECT [ID], [Name], [Group], [Price], [Description], [Duration] FROM [Treatments];",
		"INSERT INTO [Treatments] ([Name], [Group], [Price], [Description], [Duration]) VALUES (?, ?, ?, ?, ?);",
		"UPDATE [Treatments] SET [Name] = ?, [Group] = ?, [Price] = ?, [Description] = ?, [Duration] = ? WHERE [ID] = ?;",
		"DELETE FROM [Treatments] WHERE [ID] = ?;",

		// Bookings

		"SELECT [ID], [Date], [BlockNum], [TotalBlocks], [TreatmentID], [Name], [EmailAddress], [PhoneNumber], [OrderID] FROM [Bookings] WHERE [Date] = ?;",
		"INSERT INTO [Treatments] ([Date], [BlockNum], [TotalBlocks], [TreatmentID], [Name], [EmailAddress], [PhoneNumber], [OrderID]) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
		"UPDATE [Treatments] SET [Date] = ?, [BlockNum] = ?, [TotalBlocks] = ?, [TreatmentID] = ?, [Name] = ?, [EmailAddress] = ?, [PhoneNumber] = ? WHERE [ID] = ?;",
		"DELETE FROM [Treatments] WHERE [ID] = ?;",
	} {
		stmt, err := db.Prepare(ps)
		if err != nil {
			return nil, err
		}
		statements[n] = stmt
	}
	return a, nil
}

const hex = "0123456789abcdef"

func appendString(p []byte, s string) []byte {
	last := 0
	var char byte
	p = append(p, '"')
	for n, c := range s {
		switch c {
		case '"', '\\', '/':
			char = byte(c)
		case '\b':
			char = 'b'
		case '\f':
			char = 'f'
		case '\n':
			char = 'n'
		case '\r':
			char = 'r'
		case '\t':
			char = 't'
		default:
			if c < 0x20 { // control characters
				p = append(append(p, s[last:n]...), '\\', 'u', '0', '0', hex[c>>4], hex[c&0xf])
				last = n + 1
			}
			continue
		}
		p = append(append(p, s[last:n]...), '\\', char)
		last = n + 1
	}
	return append(append(p, s[last:]...), '"')
}

func appendNum(p []byte, n uint8) []byte {
	if n >= 100 {
		c := n / 100
		n -= c * 100
		p = append(p, '0'+c)
		if n < 10 {
			p = append(p, '0')
		}
	}
	if n >= 10 {
		c := n / 10
		n -= c * 10
		p = append(p, '0'+c)
	}
	return append(p, '0'+n)
}
