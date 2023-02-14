package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>URLMaid: I just cleaned this path up!  Can't you keep it clean for a day!</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Contact Page</h1><p>To get in touch, email me at <a href=\"mailto:adrutledge+b4yktg7o@gmail.com\">adrutledge+b4yktg7o@gmail.com</a>.</p>")
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<h1>FAQ!</h1>
<ul>
  <li>
    <b>Is this tool free?</b>
	  Yup!  I just wrote this to scratch an itch of mine for my own household to use.
	</li>
  <li>
    <b>Is this tool supported?</b>
		Check <a href="https://github.com/atmafox/urlmaid">github.com/atmafox/urlmaid</a> to see, but if available any support is best effort.
	  This just scratches an itch of mine.
	</li>
  <li>
    <b>How to contact me?</b>
		I'd recommend <a href="https://github.com/atmafox">github.com/atmafox</a>, but if you must have an email it's <a href="mailto:adrutledge+b4yktg7o@gmail.com">adrutledge+b4yktg7o@gmail.com</a>.
	</li>
</ul>
`)
}

func urlHandler(w http.ResponseWriter, r *http.Request) {
	urlType := chi.URLParam(r, "urlType")

	w.Write([]byte(fmt.Sprintf("URL Type to clean: %v", urlType)))
}

func main() {
	r := chi.NewRouter()

	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)
	r.Get("/url/{urlType}", urlHandler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
