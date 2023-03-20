package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"golang.org/x/crypto/acme/autocert"

	_ "github.com/atmafox/urlmaid/tidyProviders/_all"
)

var (
	flgProduction   = false
	flgRedirectHTTP = false
)

func parseFlags() {
	flag.BoolVar(&flgProduction, "production", false, "If true then start HTTPS server, generating letsencrypt cert as needed")
	flag.BoolVar(&flgRedirectHTTP, "redirect-to-https", false, "If true then redirect HTTP to HTTPS traffic")
	flag.Parse()
}

func init() {
	parseFlags()
}

type redirectHandler struct{}

func (handler redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	newURI := "https://" + r.Host + r.URL.String()
	http.Redirect(w, r, newURI, http.StatusFound)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	execTemplate(w, filepath.Join("templates", "tidy.gohtml"))
}

func execTemplate(w http.ResponseWriter, filepath string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	tpl, err := template.ParseFiles(filepath)
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
	var httpsSrv *http.Server
	var h *autocert.Manager
	certFile, keyFile := "", ""
	var rHTTP chi.Router
	var r chi.Router

	if flgProduction {
		dataDir := "."

		ha := redirectHandler{}

		r = chi.NewRouter()

		httpsSrv = &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      r,
		}
		h = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("urlmaid.ayerie.com"),
			Cache:      autocert.DirCache(dataDir),
		}
		httpsSrv.Addr = ":3001"
		httpsSrv.TLSConfig = h.TLSConfig()

		rHTTP = chi.NewRouter()
		configRouter(rHTTP)
		rHTTP.Handle("/*", h.HTTPHandler(ha))

		go func() {
			err := http.ListenAndServe(":3000", rHTTP)
			if err != nil {
				log.Fatalf("Cannot start HTTP listener: %s\n", err)
			}
		}()
	} else {
		r = chi.NewRouter()
	}

	configRouter(r)
	r.Get("/", homeHandler)

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Mount("/api", postsResource{}.Routes())
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server...")
	if flgProduction {
		err := httpsSrv.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			log.Fatalf("Listening on tcp/3001 for HTTPS failed: %s\n", err)
		}
	} else {
		err := http.ListenAndServe(":3000", r)
		if err != nil {
			log.Fatalf("Listening on tcp/3000 for HTTP failed: %s\n", err)
		}
	}
}
