package handlers

import "net/http"

const RootPath = "/"

type RootHandler struct{}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Redirect to www.luisciber.com
	w.Header().Set("Location", "https://www.luisciber.com")
	w.WriteHeader(http.StatusMovedPermanently)
}
