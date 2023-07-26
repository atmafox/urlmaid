package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/peterbourgon/ff/v3"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"golang.org/x/crypto/acme/autocert"

	_ "github.com/atmafox/urlmaid/tidyProviders/_all"
)

type redirectHandler struct {
	destPort int
	srcPort  int
}

// TODO: Make sure this works when you're going through a proxy and have to
// specify ports to the proxy itself.
func (h *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	destHost, _ := strings.CutSuffix(r.Host, fmt.Sprintf(":%d", h.srcPort))

	if h.destPort != 443 {
		destHost = fmt.Sprintf("%s:%d", destHost, h.destPort)
	}

	newURL := fmt.Sprintf("https://%s%s", destHost, r.URL.String())
	http.Redirect(w, r, newURL, http.StatusFound)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tpl, err := template.ParseFiles(filepath.Join("templates", "tidy.gohtml"))
	if err != nil {
		log.Printf("parsing template: %v", err)
		http.Error(w, "There was an error generating the page.", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error generating the page.", http.StatusInternalServerError)
		return
	}
}

func configRouter(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
}

func main() {
	flags := flag.NewFlagSet("urlmaid", flag.ContinueOnError)

	var (
		flagProduction = flags.Bool("prod", false, "Enable production mode")
		flagIp         = flags.String("ip", "", "Listen on IP address (any by default)")
		flagPort       = flags.Int("port", 3000, "Listen on port (optional)")
		flagHttpPort   = flags.Int("http-port", 3001, "Listen on http port (optional)")
		flagCertFile   = flags.String("cert-file", "", "Use cert file (use LetsEncrypt if not specified)")
		flagKeyFile    = flags.String("key-file", "", "Use key file (use LetsEncrypt if not specified)")
		flagCertDir    = flags.String("cert-dir", "", "Directory to store LetsEncrypt certificates in (default $PWD)")
		flagDebug      = flags.Bool("debug", false, "Enable debug mode")
		_              = flags.String("config", "", "config file (optional)")
	)

	err := ff.Parse(flags, os.Args[1:],
		ff.WithEnvVarPrefix("URLMAID"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
	)
	if err != nil {
		// TODO: Handle invalid env variables and config files here
		os.Exit(1)
	}

	if *flagDebug {
		log.Print("  FLAGS:")
		log.Printf("    prod:              %v\n", *flagProduction)
		log.Printf("    ip:                %q\n", *flagIp)
		log.Printf("    port:              %d\n", *flagPort)
		log.Printf("    http-port:         %d\n", *flagHttpPort)
		log.Printf("    cert-file:         %q\n", *flagCertFile)
		log.Printf("    key-file:          %q\n", *flagKeyFile)
		log.Printf("    cert-dir:          %q\n", *flagCertDir)
		log.Printf("    debug:             %v\n", *flagDebug)
	}

	// TODO: Maybe support hostnames?
	var ipAddr string
	if *flagIp != "" {
		netAddr := net.ParseIP(*flagIp)
		if netAddr == nil {
			log.Fatalf("Configuration error: IP address invalid: %q", *flagIp)
		}

		ipAddr = netAddr.String()
	}

	if *flagCertFile != "" || *flagKeyFile != "" {
		if *flagCertFile != "" && *flagKeyFile != "" {
			_, err := os.Stat(*flagCertFile)
			if err != nil {
				log.Fatalf("Configuration error: cert-file %q cannot be accessed: %q", *flagCertFile, err)
			}

			_, err = os.Stat(*flagKeyFile)
			if err != nil {
				log.Fatalf("Configuration error: key-file %q cannot be accessed: %q", *flagKeyFile, err)
			}

			if *flagDebug {
				log.Print("cert-file and key-file specified, disabling LetsEncrypt")
			}
		} else {
			log.Fatal("Configuration error: only one of cert-file or key-file specified")
		}
	}

	certDir := *flagCertDir
	if *flagCertFile == "" && *flagKeyFile == "" && certDir == "" {
		certDir = "."

		if *flagDebug {
			log.Print("None of cert-dir, cert-file, or key-file specified assuming cert-dir is $PWD")
		}
	}

	listenHttps := fmt.Sprintf("%s:%d", ipAddr, *flagPort)
	listenHttp := fmt.Sprintf("%s:%d", ipAddr, *flagHttpPort)

	r := chi.NewRouter()
	var httpsSrv *http.Server

	if *flagProduction {
		redirHa := &redirectHandler{
			*flagPort,
			*flagHttpPort,
		}

		httpsSrv = &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Addr:         listenHttps,
			Handler:      r,
		}

		rHttp := chi.NewRouter()
		configRouter(rHttp)

		var httpHa http.Handler
		httpHa = redirHa

		if certDir != "" {
			acManager := &autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist("urlmaid.ayerie.com"),
				Cache:      autocert.DirCache(certDir),
			}
			httpHa = acManager.HTTPHandler(redirHa)

			httpsSrv.TLSConfig = acManager.TLSConfig()
		}

		rHttp.Handle("/*", httpHa)

		go func() {
			err := http.ListenAndServe(listenHttp, rHttp)
			if err != nil {
				log.Fatalf("Cannot start HTTP redirect listener: %q", err)
			}
		}()
	}

	configRouter(r)
	r.Get("/", homeHandler)

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Mount("/api", postsResource{}.Routes())
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	log.Print("Starting the server...")
	if *flagProduction {
		err := httpsSrv.ListenAndServeTLS(*flagCertFile, *flagKeyFile)
		if err != nil {
			log.Fatalf("Listening on tcp/%d for HTTPS failed: %q", *flagPort, err)
		}
	} else {
		err := http.ListenAndServe(listenHttp, r)
		if err != nil {
			log.Fatalf("Listening on tcp/%d for HTTP failed: %q", *flagHttpPort, err)
		}
	}
}
