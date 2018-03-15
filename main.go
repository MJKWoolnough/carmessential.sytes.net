package main

import (
	"flag"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"path"
	"strings"

	"github.com/MJKWoolnough/httpbuffer"
	_ "github.com/MJKWoolnough/httpbuffer/gzip"
	"github.com/MJKWoolnough/webserver/proxy/client"
)

var (
	databaseFile = flag.String("d", "./database.db", "database file")
	filesDir     = flag.String("f", "./files", "files directory")
	logName      = flag.String("n", "", "name for logging")
	initConfig   = flag.Bool("init", false, "used for initial configuration")
	logger       *log.Logger
)

const configPrefix = "CECONFIG_"

func main() {
	flag.Parse()
	logger = log.New(os.Stderr, *logName, log.LstdFlags)

	err := DB.init(*databaseFile)
	if err != nil {
		logger.Printf("error while opening database: %s\n", err)
		return
	}
	if *initConfig {
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, configPrefix) {
				parts := strings.SplitN(strings.TrimPrefix(env, configPrefix), "=", 2)
				logger.Printf("CONFIG: Setting %q to %q\n", parts[0], parts[1])
				Config.Set(parts[0], parts[1])
			}
		}
	}
	Email.init(Config.Get("emailSMTP"), Config.Get("emailLogin"), smtp.PlainAuth("", Config.Get("emailLogin"), Config.Get("emailPassword"), Config.Get("emailHost")))
	Session.init(Config.Get("sessionKey"), Config.Get("basketKey"))
	BasketInit(*filesDir)
	err = Pages.init(path.Join(*filesDir, "template.tmpl"))
	if err != nil {
		log.Printf("error while opening templates: %s\n", err)
		return
	}

	err = User.init(
		path.Join(*filesDir, "login.tmpl"),
		path.Join(*filesDir, "register.tmpl"),
		path.Join(*filesDir, "email.tmpl"),
		Config.Get("emailFrom"),
		Config.Get("registrationKey"),
	)
	if err != nil {
		log.Printf("error while opening user templates: %s\n", err)
		return
	}
	Admin.init()

	// load items from database
	// load schedule from database
	wrapped := http.NewServeMux()
	/*
		wrapped.Handle("/treatments/", Treatments)
		wrapped.Handle("/vouchers/", Vouchers)
		wrapped.Handle("/contact.html", contact)
		wrapped.Handle("/pricelist.html", Treatments.PriceList)
		wrapped.Handle("/user/", user)
	*/
	wrapped.Handle("/admin/", &Admin)
	wrapped.Handle("/user/", &User)
	wrapped.Handle("/login.html", http.HandlerFunc(User.Login))
	wrapped.Handle("/logout.html", http.HandlerFunc(User.Logout))
	wrapped.Handle("/register.html", http.HandlerFunc(User.Register))
	wrapped.Handle("/terms.html", NewPageFile("CARMEssential - Terms &amp; Conditions", "terms", "", path.Join(*filesDir, "terms.html"), true))
	wrapped.Handle("/about.html", NewPageFile("CARMEssential - About Me", "about", "", path.Join(*filesDir, "about.html"), true))
	wrapped.Handle("/", NewPageFile("CARMEssential", "home", "", path.Join(*filesDir, "index.html"), true))
	http.Handle("/assets/", http.FileServer(http.Dir(*filesDir)))
	//http.Handle("/checkout.html", Pages.SemiWrap(basket))
	http.Handle("/", httpbuffer.Handler{wrapped})

	cc := make(chan struct{})
	go func() {
		logger.Println("Server Started")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, os.Interrupt)
		select {
		case <-sc:
			logger.Println("Closing")
		case <-cc:
		}
		signal.Stop(sc)
		close(sc)
		client.Close()
		client.Wait()
		close(cc)
	}()

	err = client.Run()

	select {
	case <-cc:
	default:
		logger.Println(err)
		cc <- struct{}{}
	}
	<-cc
}
