package main // import "vimagination.zapto.org/carmessential.sytes.net"

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"vimagination.zapto.org/httpbuffer"
	_ "vimagination.zapto.org/httpbuffer/gzip"
	"vimagination.zapto.org/webserver/proxy/client"
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

	if err := Pages.Init(); err != nil {
		logger.Printf("error initialising pages: %s\n", err)
	}
	if err := DB.Init(); err != nil {
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
	for _, init := range [...]func() error{
		Email.Init,
		Session.Init,
		BasketInit,
		User.Init,
		Admin.Init,
		Contact.Init,
	} {
		if err := init(); err != nil {
			log.Printf("error during initialisation: %s\n", err)
			return
		}
	}

	// load items from database
	// load schedule from database
	wrapped := http.NewServeMux()
	wrapped.Handle("/treatments.html", &Treatments)
	wrapped.Handle("/contact.html", &Contact)
	/*
		wrapped.Handle("/vouchers/", Vouchers)
		wrapped.Handle("/pricelist.html", Treatments.PriceList)
		wrapped.Handle("/user/", user)
	*/
	wrapped.Handle("/admin/", &Admin)
	wrapped.Handle("/user/", &User)
	wrapped.Handle("/login.html", http.HandlerFunc(User.Login))
	wrapped.Handle("/logout.html", http.HandlerFunc(User.Logout))
	wrapped.Handle("/register.html", http.HandlerFunc(User.Register))
	wrapped.Handle("/terms.html", NewPageFile("CARMEssential - Terms & Conditions", "terms", "terms.html"))
	wrapped.Handle("/about.html", NewPageFile("CARMEssential - About Me", "about", "about.html"))
	wrapped.Handle("/", NewPageFile("CARMEssential", "default", "index.html"))
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
