package main

import (
	"io/fs"
	"net/http"
	"strings"

	frontend "github.com/Tattsum/translate-prompt/frontend"
)

func registerSPA(mux *http.ServeMux) {
	dist, err := fs.Sub(frontend.Dist, "dist")
	if err != nil {
		return
	}
	fileServer := http.FileServer(http.FS(dist))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if isAPIPath(r.URL.Path) {
			http.NotFound(w, r)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(dist, path); err != nil {
			r.URL.Path = "/index.html"
		}
		fileServer.ServeHTTP(w, r)
	})
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/query") ||
		strings.HasPrefix(path, "/playground") ||
		strings.HasPrefix(path, "/translate_prompt.v1.TranslatePromptService/")
}
