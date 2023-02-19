package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

type CleanedURL struct {
	Type    template.HTML
	Input   template.HTML
	Cleaned template.HTML
}

type URLToEncode struct {
	URL  string
	Type string
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
	fmt.Fprint(w, `ebay
amazon`)
}

func tidyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	t := chi.URLParam(r, "type")
	u := chi.URLParam(r, "url")

	// TODO: Verify sanity of request

	doTidy(t, u, w)
}

func tidyPost(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	var u URLToEncode
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	doTidy(u.Type, url.QueryEscape(u.URL), w)
}

func doTidy(t string, u string, w http.ResponseWriter) {
	switch t {
	case "autodetect":
		// Figure out how to autodetect
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	case "ebay":
		d, err := url.QueryUnescape(u)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		out := strings.Split(d, "?")[0]

		fmt.Fprintf(w, "%s", out)
		return
	case "amazon":
		d, err := url.QueryUnescape(u)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		r, err := regexp.Compile(`(?P<useful>/dp/[[:alnum:]]+)/`)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		match := r.FindStringSubmatch(d)
		result := make(map[string]string)

		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		fmt.Fprintf(w, "https://amazon.com%s", result["useful"])
		return
	case "default":
		// TODO: Perhaps a different error code is better for an API?  Research.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
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
	r.Post("/api/v1/tidy", tidyPost)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
