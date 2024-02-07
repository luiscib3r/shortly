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
	query := r.URL.Query()

	if shortcut, ok := h.repository.FindById(id); ok {
		redirectTo := shortcut.Url() + "?" + query.Encode()
		w.Header().Set("Location", redirectTo)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.NotFound(w, r)
	}

}
