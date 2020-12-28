package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8780"
	}
	log.Printf("listening on port %q", port)
	http.ListenAndServe(":"+port, http.HandlerFunc(handler))
}

// handler handles Go vanity import requests
func handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	w.Header().Set("Cache-Control", "public")
	vanity.Execute(w, path)
}

var vanity = template.Must(template.New("vanity").Parse(vanityTemplate))

var vanityTemplate = `<html>
  <head>
    <meta name="go-import" content="cli.gofig.dev git https://github.com/gofigdev/cli">
  </head>
  <body>
    Install: go get cli.gofig.dev/{{ . }}@latest <br>
    <a href="http://pkg.go.dev/src.gofig.dev/{{ . }}">Documentation</a><br>
  </body>
</html>
`
