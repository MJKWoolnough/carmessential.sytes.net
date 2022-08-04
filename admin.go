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
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/websocket"
	"vimagination.zapto.org/form"
	"vimagination.zapto.org/jsonrpc"
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

type booking struct {
	ID          uint64 `json:"id"`
	Date        uint64 `json:"date"`
	BlockNum    uint8  `json:"blockNum"`
	TotalBlocks uint8  `json:"totalBlock"`
	TreatmentID uint64 `json:"treatmentID"`
	Name        string `json:"name"`
	Email       string `json:"emailAddress"`
	Phone       string `json:"phoneNumber"`
	OrderID     uint64 `json:"orderID"`
}

type voucher struct {
	ID        uint64 `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Expiry    uint64 `json:"expiry"`
	OrderID   uint64 `json:"orderID"`
	IsValue   bool   `json:"isValue"`
	Value     uint64 `json:"value"`
	Valid     bool   `json:"valid"`
	OrderUsed uint64 `json:"orderUsed"`
}

type order struct {
	Name         string    `json:"name"`
	Price        uint64    `json:"price"`
	Bookings     []booking `json:"bookings"`
	Vouchers     []voucher `json:"vouchers"`
	UsedVouchers []uint64  `json:"usedVouchers"`
}

const (
	setHeaderFooter = iota

	listTreatments
	addTreatment
	setTreatment
	removeTreatment

	orderTime
	getOrders
	addOrder
	removeOrder
	removeOrderBookings
	removeOrderVouchers

	listBookings
	addBooking
	updateBooking
	removeBooking

	getVoucher
	getVoucherByCode
	addVoucher
	updateVoucher
	removeVoucher
	setVoucherValid
	checkVoucherCode
	useVoucher

	totalStmts
)

const codeChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

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
	fmt.Fprintf(wconn, "{\"id\":-2,\"result\":[%q,%q]}", header, footer)
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
		var t treatment
		buf := json.RawMessage{'['}
		first := true
		for r.Next() {
			if err := r.Scan(&t.ID, &t.Name, &t.Group, &t.Price, &t.Description, &t.Duration); err != nil {
				return nil, err
			}
			if first {
				first = false
			} else {
				buf = append(buf, ',')
			}
			buf = strconv.AppendUint(append(buf, "{\"id\":"...), t.ID, 10)
			buf = appendString(append(buf, ",\"name\":"...), t.Name)
			buf = appendString(append(buf, ",\"group\":"...), t.Group)
			buf = strconv.AppendUint(append(buf, ",\"price\":"...), uint64(t.Price), 10)
			buf = appendString(append(buf, ",\"description\":"...), t.Description)
			buf = strconv.AppendUint(append(buf, ",\"price\":"...), uint64(t.Duration), 10)
			buf = append(buf, '}')
		}
		return append(buf, ']'), nil
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
		var id uint64
		if err := json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		if _, err := statements[removeTreatment].Exec(id); err != nil {
			return nil, err
		}
		// remove treatment page
		generatePages(-1)
		return nil, nil
	case "getOrderTime":
		var id uint64
		if err := json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		var t uint64
		if err := statements[orderTime].QueryRow(id).Scan(&t); err != nil {
			return nil, err
		}
		return t, nil
	case "addOrder":
		var order order
		if err := json.Unmarshal(data, &order); err != nil {
			return nil, err
		}
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()
		r, err := tx.Stmt(statements[addOrder]).Exec(uint64(time.Now().Unix()), order.Name, order.Price)
		if err != nil {
			return nil, err
		}
		oid, err := r.LastInsertId()
		if err != nil {
			return nil, err
		}
		orderID := uint64(oid)
		buf := strconv.AppendUint(append(data[:0], "{\"orderID\":"...), orderID, 10)
		buf = append(buf, ",\"bookings\":["...)
		if len(order.Bookings) > 0 {
			addBooking := tx.Stmt(statements[addBooking])
			for n, b := range order.Bookings {
				r, err := addBooking.Exec(b.Date, b.BlockNum, b.TotalBlocks, b.TreatmentID, b.Name, b.Email, b.Phone, orderID)
				if err != nil {
					return nil, err
				}
				id, err := r.LastInsertId()
				if err != nil {
					return nil, err
				}
				if n > 0 {
					buf = append(buf, ',')
				}
				buf = strconv.AppendUint(buf, uint64(id), 10)
			}
		}
		buf = append(buf, "],\"vouchers\":["...)
		if len(order.Vouchers) > 0 {
			addVoucher := tx.Stmt(statements[addVoucher])
			checkVoucher := tx.Stmt(statements[checkVoucherCode])
			for n, v := range order.Vouchers {
				code := make([]byte, 0, 10)
				for valid := 1; valid == 1; {
					code = code[:8+rand.Intn(3)]
					for n := range code {
						code[n] = codeChars[rand.Intn(len(codeChars))]
					}
					v.Code = string(code)
					if err := checkVoucher.QueryRow(v.Code).Scan(&valid); err != nil {
						return nil, err
					}
				}
				r, err := addVoucher.Exec(v.Code, v.Name, v.Expiry, v.OrderID, v.IsValue, v.Valid, v.Valid)
				if err != nil {
					return nil, err
				}
				id, err := r.LastInsertId()
				if err != nil {
					return nil, err
				}
				if n > 0 {
					buf = append(buf, ',')
				}
				buf = strconv.AppendUint(buf, uint64(id), 10)
			}
		}
		buf = append(buf, ']', '}')
		if len(order.UsedVouchers) > 0 {
			useVoucher := tx.Stmt(statements[useVoucher])
			for _, u := range order.UsedVouchers {
				if _, err := useVoucher.Exec(orderID, u); err != nil {
					return nil, err
				}
			}
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return buf, nil
	case "removeOrder":
		var id uint64
		if err := json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()
		if _, err := tx.Stmt(statements[removeOrderBookings]).Exec(id); err != nil {
			return nil, err
		}
		if _, err := tx.Stmt(statements[removeOrderVouchers]).Exec(id); err != nil {
			return nil, err
		}
		if _, err := tx.Stmt(statements[removeOrder]).Exec(id); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return nil, nil
	case "listBookings":
		var dates [2]uint64
		if err := json.Unmarshal(data, &dates); err != nil {
			return nil, err
		}
		r, err := statements[listBookings].Query(dates[0], dates[1])
		if err != nil {
			return nil, err
		}
		var b booking
		buf := json.RawMessage{'['}
		first := true
		for r.Next() {
			if err := r.Scan(&b.ID, &b.Date, &b.BlockNum, &b.TotalBlocks, &b.TreatmentID, &b.Name, &b.Email, &b.Phone, &b.OrderID); err != nil {
				return nil, err
			}
			if first {
				first = false
			} else {
				buf = append(buf, ',')
			}
			buf = strconv.AppendUint(append(buf, "{\"id\":"...), b.ID, 10)
			buf = strconv.AppendUint(append(buf, ",\"date\":"...), b.Date, 10)
			buf = appendNum(append(buf, ",\"blockNum\":"...), b.BlockNum)
			buf = appendNum(append(buf, ",\"totalBlocks\":"...), b.TotalBlocks)
			buf = strconv.AppendUint(append(buf, ",\"treatmentID\":"...), b.TreatmentID, 10)
			buf = appendString(append(buf, ",\"name\":"...), b.Name)
			buf = appendString(append(buf, ",\"emailAddress\":"...), b.Email)
			buf = appendString(append(buf, ",\"phoneNumber\":"...), b.Phone)
			buf = strconv.AppendUint(append(buf, ",\"orderID\":"...), b.OrderID, 10)
			buf = append(buf, '}')
		}
		return append(buf, ']'), nil
	case "updateBooking":
		var b booking
		if err := json.Unmarshal(data, &b); err != nil {
			return nil, err
		}
		if _, err := statements[updateBooking].Exec(b.Date, b.BlockNum, b.TotalBlocks, b.TreatmentID, b.Name, b.Email, b.Phone); err != nil {
			return nil, err
		}
		return nil, nil
	case "removeBooking":
		var bID uint64
		if err := json.Unmarshal(data, &bID); err != nil {
			return nil, err
		}
		if _, err := statements[removeBooking].Exec(bID); err != nil {
			return nil, err
		}
		return nil, nil
	case "getVoucher":
		var id uint64
		if err := json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		var (
			v       voucher
			isValue uint8
			valid   uint8
		)
		if err := statements[getVoucher].QueryRow(id).Scan(&v.Code, &v.Name, &v.Expiry, &v.OrderID, isValue, &v.Value, &valid, &v.OrderUsed); err != nil {
			return nil, err
		}
		v.ID = id
		v.IsValue = isValue == 1
		v.Valid = valid == 1
		return v, nil
	case "getVoucherByCode":
		var code string
		if err := json.Unmarshal(data, &code); err != nil {
			return nil, err
		}
		var (
			v       voucher
			isValue uint8
			valid   uint8
		)
		if err := statements[getVoucherByCode].QueryRow(code).Scan(&v.ID, &v.Name, &v.Expiry, &v.OrderID, isValue, &v.Value, &valid, &v.OrderUsed); err != nil {
			return nil, err
		}
		v.Code = code
		v.IsValue = isValue == 1
		v.Valid = valid == 1
		return v, nil
	case "updateVoucher":
		var v voucher
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		if _, err := statements[updateVoucher].Exec(v.Name, v.Expiry, v.ID); err != nil {
			return nil, err
		}
		return nil, nil
	case "removeVoucher":
		var id uint64
		if err := json.Unmarshal(data, &id); err != nil {
			return nil, err
		}
		if _, err := statements[updateVoucher].Exec(id); err != nil {
			return nil, err
		}
		return nil, nil
	case "setVoucherValid":
		var idValid struct {
			ID    uint64 `json:"id"`
			Valid bool   `json:"valid"`
		}
		if err := json.Unmarshal(data, &idValid); err != nil {
			return nil, err
		}
		valid := 0
		if idValid.Valid {
			valid = 1
		}
		if _, err := statements[setVoucherValid].Exec(valid, idValid.ID); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, errors.New("unknown endpoint")
	}
}

func generatePages(id int64) {
}

func init() {
	if a, err := adminInit(); err == nil {
		http.Handle("/admin", a)
	} else {
		fmt.Fprintln(os.Stderr, err)
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
		"[Treatments]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Name] TEXT NOT NULL, [Group] TEXT NOT NULL DEFAULT '', [Price] INTEGER NOT NULL, [Description] TEXT NOT NULL DEFAULT '', [Duration] INTEGER NOT NULL, [Deleted] BOOLEAN DEFAULT 1 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Orders]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Time] INTEGER NOT NULL, [Name] TEXT NOT NULL, [Total] INTEGER NOT NULL, [Deleted] BOOLEAN DEFAULT 1 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Bookings]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [Date] INTEGER NOT NULL, [BlockNum] INTEGER NOT NULL, [TotalBlocks] INTEGER NOT NULL, [TreatmentID] INTEGER NOT NULL, [Name] TEXT NOT NULL DEFAULT '', [EmailAddress] NOT NULL DEFAULT '', [PhoneNumber] NOT NULL DEFAULT '', [OrderID] INTEGER NOT NULL, [Deleted] BOOLEAN DEFAULT 1 NOT NULL CHECK ([Deleted] IN (0,1)));",
		"[Vouchers]([ID] INTEGER PRIMARY KEY AUTOINCREMENT, [CODE] TEXT NOT NULL UNIQUE, [Name] TEXT NOT NULL, [Expiry] INTEGER NOT NULL, [OrderID] INTEGER NOT NULL, [IsValue] BOOLEAN DEFAULT 0 NOT NULL CHECK ([IsValue] IN (0,1)), [Value] INTEGER NOT NULL, [Valid] BOOLEAN DEFAULT 1 NOT NULL CHECK ([Valid] IN (0,1)), [OrderUsed] INTEGER NOT NULL DEFAULT 0, [Deleted] BOOLEAN DEFAULT 1 NOT NULL CHECK ([Deleted] IN (0,1)));",
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
	} else if err = db.QueryRow("SELECT [Header], [Footer] FROM [Settings];").Scan(&header, &footer); err != nil {
		return nil, err
	}
	for n, ps := range []string{
		"UPDATE [Settings] SET [Header] = ?, [Footer] = ?;",

		// Treatments
		"SELECT [ID], [Name], [Group], [Price], [Description], [Duration] FROM [Treatments];",
		"INSERT INTO [Treatments] ([Name], [Group], [Price], [Description], [Duration]) VALUES (?, ?, ?, ?, ?);",
		"UPDATE [Treatments] SET [Name] = ?, [Group] = ?, [Price] = ?, [Description] = ?, [Duration] = ? WHERE [ID] = ?;",
		"UPDATE [Treatments] SET [Deleted] = 1 WHERE [ID] = ?;",

		// Orders
		"SELECT [Time] FROM [Orders] WHERE [ID] = ?;",
		"SELECT [Time], [Name], [Total] FROM [Orders];",
		"INSERT INTO [Orders] ([Time], [Name], [Total]) VALUES (?, ?, ?);",
		"UPDATE [Orders] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Bookings] SET [Deleted] = 1 WHERE [OrderID] = ?;",
		"UPDATE [Vouchers] SET [Deleted] = 1 WHERE [OrderID] = ?;",

		// Bookings
		"SELECT [ID], [Date], [BlockNum], [TotalBlocks], [TreatmentID], [Name], [EmailAddress], [PhoneNumber], [OrderID] FROM [Bookings] WHERE [Date] BETWEEN ? AND ? ORDER BY [Date] ASC, [BlockNum] ASC;",
		"INSERT INTO [Bookings] ([Date], [BlockNum], [TotalBlocks], [TreatmentID], [Name], [EmailAddress], [PhoneNumber], [OrderID]) VALUES (?, ?, ?, ?, ?, ?, ?, ?);",
		"UPDATE [Bookings] SET [Date] = ?, [BlockNum] = ?, [TotalBlocks] = ?, [TreatmentID] = ?, [Name] = ?, [EmailAddress] = ?, [PhoneNumber] = ? WHERE [ID] = ?;",
		"UPDATE [Bookings] Set [Deleted] = 1 WHERE [ID] = ?;",

		// Vouchers
		"SELECT [Code], [Name], [Expiry], [OrderID], [IsValue], [Value], [Valid], [OrderUsed] FROM [Vouchers] WHERE [ID] = ?;",
		"SELECT [ID], [Name], [Expiry], [OrderID], [IsValue], [Value], [Valid], [OrderUsed] FROM [Vouchers] WHERE [Code] = ?;",
		"INSERT INTO [Vouchers] ([Code], [Name], [Expiry], [OrderID], [IsValue], [Value]) VALUES (?, ?, ?, ?, ?, ?);",
		"UPDATE [Vouchers] SET [Name] = ?, [Expiry] = ? WHERE [ID] = ?;",
		"UPDATE [Vouchers] SET [Deleted] = 1 WHERE [ID] = ?;",
		"UPDATE [Vouchers] SET [Valid] = ? WHERE [ID] = ?;",
		"UPDATE [Vouchers] SET [Valid] = 0, [OrderUsed] = ? WHERE [ID] = ?;",
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
