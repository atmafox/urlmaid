package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/atmafox/urlmaid/tidyProviders"
	"github.com/go-chi/chi/v5"
)

type URLToEncode struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type postsResource struct{}

// Let's set up our router for the API
func (rs postsResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/v1/supported", rs.Supp)
	r.Post("/v1/tidy", rs.Tidy)
	r.Get("/supported", rs.Supp)
	r.Post("/tidy", rs.Tidy)

	return r
}

// This one is unique in that it outputs plain and only uses a template to make things easier to update
func (rs postsResource) Supp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	a := r.Header["Accept"][0]
	tidiers := tidyProviders.Tidiers
	n := make([]string, 0, len(tidiers)+1)

	n = append(n, "autodetect")
	for t := range tidiers {
		n = append(n, t)
	}

	switch {
	case strings.HasPrefix(a, "text/plain"):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		var o string

		// TODO: Generate the list of supported types 'automatically'
		for i := range n {
			o = fmt.Sprintf("%s\n", n[i])
			fmt.Fprintf(w, "%s", o)
		}
		return
	case strings.HasPrefix(a, "application/json"):
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		e := json.NewEncoder(w)

		err := e.Encode(n)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	default:
		http.Error(w, "Bad reqeust", http.StatusBadRequest)
		return
	}
}

func (rs postsResource) Tidy(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	enc := json.NewEncoder(w)

	defer r.Body.Close()

	var u URLToEncode
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	out := doTidy(u.Type, u.URL, w)

	a := r.Header["Accept"][0]
	switch {
	case strings.HasPrefix(a, "text/plain"):
		fmt.Fprintf(w, "%s", out)

	case strings.HasPrefix(a, "application/json"):
		err := enc.Encode(out)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

}

func doTidy(t string, u string, w http.ResponseWriter) string {
	tidiers := tidyProviders.Tidiers

	if len(tidiers) == 0 {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return ""
	}

	f := func() bool {
		b, err := tidiers[t].GetURLMatch(u)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return b
	}

	switch {
	case t == "autodetect":
		// Run each registered detector
		for tidier := range tidiers {
			if match, _ := tidiers[tidier].GetURLMatch(u); match {
				out, err := tidiers[tidier].TidyURL(u)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return ""
				}

				return out
			}
		}
	case f():
		out, err := tidiers[t].TidyURL(u)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return ""
		}
		return out
	case false:
		// TODO: Perhaps a different error code is better for an API?  Research.
		http.Error(w, "Bad request", http.StatusBadRequest)
		return ""
	}

	return ""
}
