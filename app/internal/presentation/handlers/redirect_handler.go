package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luiscib3r/shortly/app/internal/domain/repositories"
)

type RedirectHandler struct {
	repository repositories.ShortcutRepository
}

func NewRedirectHandler(
	repository repositories.ShortcutRepository,
) *RedirectHandler {
	return &RedirectHandler{
		repository: repository,
	}
}

func (h *RedirectHandler) Redirect(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	id := vars["id"]
	query := r.URL.Query().Encode()

	if shortcut, ok := h.repository.FindById(id); ok {
		var redirectTo string
		if query == "" {
			redirectTo = shortcut.Url()
		} else {
			redirectTo = shortcut.Url() + "?" + query
		}
		w.Header().Set("Location", redirectTo)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.NotFound(w, r)
	}

}
