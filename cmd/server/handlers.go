package server

import (
	"net/http"

	"github.com/pchchv/sws/helpers/ancli"
)

func slogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ancli.PrintfOK("%s - %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func cacheHandler(next http.Handler, cacheControl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", cacheControl)
		next.ServeHTTP(w, r)
	})
}
