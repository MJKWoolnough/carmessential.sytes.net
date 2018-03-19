package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

	if err := DB.init(); err != nil {
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
	if err := Email.init(); err != nil {
		log.Printf("error initialising Email: %s\n", err)
		return
	}
	if err := Session.init(); err != nil {
		log.Printf("error initialising Sessions: %s\n", err)
		return
	}
	if err := Pages.init(); err != nil {
		log.Printf("error while opening templates: %s\n", err)
		return
	}
	if err := BasketInit(); err != nil {
		logger.Printf("error initialising Basket: %s\n", err)
		return
	}

	if err := User.init(); err != nil {
		log.Printf("error while opening user templates: %s\n", err)
		return
	}
	if err := Admin.init(); err != nil {
		logger.Printf("error initialising Admin: %s\n", err)
		return
	}

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
	wrapped.Handle("/terms.html", NewPageFile("CARMEssential - Terms &amp; Conditions", "terms", "", filepath.Join(*filesDir, "terms.html"), true))
	wrapped.Handle("/about.html", NewPageFile("CARMEssential - About Me", "about", "", filepath.Join(*filesDir, "about.html"), true))
	wrapped.Handle("/", NewPageFile("CARMEssential", "home", "", filepath.Join(*filesDir, "index.html"), true))
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

	err := client.Run()

	select {
	case <-cc:
	default:
		logger.Println(err)
		cc <- struct{}{}
	}
	<-cc
}
