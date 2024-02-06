package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luiscib3r/shortly/app/internal/domain/repositories"
	"github.com/luiscib3r/shortly/app/internal/presentation/dtos"
)

type ShortcutHandler struct {
	repository  repositories.ShortcutRepository
	environment repositories.EnvironmentRepository
}

func NewShortcutHandler(
	repository repositories.ShortcutRepository,
	environment repositories.EnvironmentRepository,
) *ShortcutHandler {
	return &ShortcutHandler{
		repository:  repository,
		environment: environment,
	}
}

func (h *ShortcutHandler) FindAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	shortcuts := h.repository.FindAll()

	response := make([]dtos.ShortcutDto, len(shortcuts))

	for i, shortcut := range shortcuts {
		response[i] = dtos.ShortcutDto{
			Id:    shortcut.Id(),
			Url:   shortcut.Url(),
			Short: h.getShortUrl(shortcut.Id()),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ShortcutHandler) Save(
	w http.ResponseWriter,
	r *http.Request,
) {
	var payload dtos.CreateShortcutDto

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortcut := h.repository.SaveUrl(payload.Url)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dtos.ShortcutDto{
		Id:    shortcut.Id(),
		Url:   shortcut.Url(),
		Short: h.getShortUrl(shortcut.Id()),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ShortcutHandler) FindById(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]

	if shortcut, ok := h.repository.FindById(id); ok {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(dtos.ShortcutDto{
			Id:    shortcut.Id(),
			Url:   shortcut.Url(),
			Short: h.getShortUrl(shortcut.Id()),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *ShortcutHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]

	if ok := h.repository.Delete(id); ok {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *ShortcutHandler) getShortUrl(id string) string {
	return h.environment.GetBaseUrl() + "/" + id
}
