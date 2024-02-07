package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/luiscib3r/shortly/app/internal/domain/repositories"
	"github.com/luiscib3r/shortly/app/internal/presentation/dtos"
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
		original, err := url.ParseRequestURI(shortcut.Url())

		if err != nil {
			response := dtos.ErrorDto{
				Message: "Failed to parse original URL in shortcut",
				Error:   err.Error(),
			}

			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}

		var redirectTo string
		if query == "" {
			redirectTo = original.String()
		} else if original.RawQuery == "" {
			redirectTo = original.String() + "?" + query
		} else {
			redirectTo = original.String() + "&" + query
		}
		w.Header().Set("Location", redirectTo)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.NotFound(w, r)
	}

}
