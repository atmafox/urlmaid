package main

import (
	"context"
	"crypto/tls"
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	execTemplate(w, filepath.Join("templates", "tidy.gohtml"))
}

// Commenting the next two out until something better is written
/*
func contactHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	execTemplate(w, filepath.Join("templates", "contact.gohtml"))
}
*/

/*
func faqHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	execTemplate(w, filepath.Join("templates", "faq.gohtml"))
}
*/

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

// demo code, rewrite later to make more sense
func makeHTTPServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", homeHandler)

	// set timeouts so that a slow or malicious client doesn't
	// hold resources forever
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}
}

func main() {
	var httpsSrv *http.Server
	var m *autocert.Handler

	if inProduction {
		dataDir := "."
		hostPolicy := func(ctx context.Context, host string) error {
			// make this a commandline option
			allowedHost := "urlmaid.ayerie.com"
			if host == allowedHost {
				return nil
			}
			return fmt.Errorf("acme/autocert: only %s host is allowed", allowedHost)
		}

		httpsSrv = makeHTTPServer()
		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(dataDir),
		}
		httpsSrv.Addr = ":3001"
		httpsSrv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		go func() {
			err := httpsSrv.ListenAndServeTLS("", "")
			if err != nil {
				log.Fatalf("httpsSrv.ListendAndServeTLS() failed with %s", err)
			}
		}()
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", homeHandler)
	// Just getting rid of these for now
	//r.Get("/contact", contactHandler)
	//r.Get("/faq", faqHandler)
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Mount("/api", postsResource{}.Routes())
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server...")
	http.ListenAndServe(":3000", r)
	http.ListenAndServeTLS(":3001", r)
}
