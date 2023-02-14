package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type CleanedURL struct {
	Type    template.HTML
	Input   template.HTML
	Cleaned template.HTML
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	execTemplate(w, filepath.Join("templates", "home.gohtml"))
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	execTemplate(w, filepath.Join("templates", "contact.gohtml"))
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	execTemplate(w, filepath.Join("templates", "faq.gohtml"))
}

func urlHandler(w http.ResponseWriter, r *http.Request) {
	execTemplate(w, filepath.Join("templates", "url.gohtml"))
}

// This one is unique in that it outputs plain and only uses a template to make things easier to update
func suppHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// TODO: Generate the list of supported types 'automatically'
	fmt.Fprint(w, `autodetect
ebay`)
}

func tidyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	t := chi.URLParam(r, "type")
	u := chi.URLParam(r, "url")

	// TODO: Verify sanity of request

	switch t {
	case "autodetect":
		// Figure out how to autodetect
	case "ebay":
		d, err := url.QueryUnescape(u)
		if err != nil {
			panic(err) // TODO: Handle errors usefully
		}

		out := strings.Split(d, "?")[0]

		fmt.Fprintf(w, "%s", out)
	}
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

func main() {
	r := chi.NewRouter()

	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)
	r.Get("/url", urlHandler)
	r.Get("/api/supported", suppHandler)
	r.Get("/api/v1/supported", suppHandler)
	r.Get("/api/tidy/{type}/{url}", tidyHandler)
	r.Get("/api/v1/tidy/{type}/{url}", tidyHandler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
