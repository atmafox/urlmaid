package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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

	switch {
	case strings.HasPrefix(a, "text/plain"):
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		// TODO: Generate the list of supported types 'automatically'
		fmt.Fprint(w, `ebay
amazon`)
		return
	case strings.HasPrefix(a, "application/json"):
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		s := [2]string{
			"ebay",
			"amazon",
		}

		e := json.NewEncoder(w)

		err := e.Encode(s)
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
	switch t {
	case "autodetect":
		// Figure out how to autodetect
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return ""
	case "ebay":
		d := u

		out := strings.Split(d, "?")[0]

		return out
	case "amazon":
		d := u

		r, err := regexp.Compile(`(?P<useful>/dp/[[:alnum:]]+)/`)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return ""
		}

		match := r.FindStringSubmatch(d)
		result := make(map[string]string)

		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}

		out := fmt.Sprintf("https://amazon.com%s", result["useful"])
		return out
	case "default":
		// TODO: Perhaps a different error code is better for an API?  Research.
		http.Error(w, "Bad request", http.StatusBadRequest)
		return ""
	}

	return ""
}
