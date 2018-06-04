package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"text/template"

	"github.com/gorilla/handlers"
)

var (
	laddr = flag.String("p", ":8080", "listen address")

	expanded = regexp.MustCompile(`(.*\.repo|metalink.xml)$`)
)

// the expander middleware expands the response of a selected
// requests (matched by the regexp) by parsing the responses
// as go templates, and passing the request as template context.
//
// The main goal is to allow the file to dynamically refer to the
// actual URL (IP) of the webserver.
type expander struct {
	pattern *regexp.Regexp
	handler http.Handler
}

func (e *expander) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e.pattern.MatchString(r.URL.Path) {
		var rec httptest.ResponseRecorder
		rec.Body = bytes.NewBuffer(nil)

		e.handler.ServeHTTP(&rec, r)
		for k, v := range rec.HeaderMap {
			// we expand the template below and thus the content length will change.
			if k != "Content-Length" {
				w.Header()[k] = v
			}
		}
		if r.URL.Path == "/metalink.xml" {
			w.Header().Set("Content-Type", "application/metalink+xml")
		}
		t, err := template.New("").Parse(rec.Body.String())
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if err := t.Execute(w, r); err != nil {
			log.Printf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		//		w.Write(rec.Body.Bytes())
		return
	}
	e.handler.ServeHTTP(w, r)
}

func main() {
	flag.Parse()
	fs := http.FileServer(http.Dir("."))
	i := &expander{pattern: expanded, handler: fs}
	l := handlers.LoggingHandler(os.Stdout, i)
	http.ListenAndServe(*laddr, l)
}
