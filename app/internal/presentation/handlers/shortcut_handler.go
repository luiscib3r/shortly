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

// @Summary	Get Shortcuts
// @Tags Shortcut
// @Produce json
// @Success 200 {array} dtos.ShortcutDto
// @Failure 500 {object} dtos.ErrorDto
// @Router /api/shortcut [get]
func (h *ShortcutHandler) FindAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")
	shortcuts, err := h.repository.FindAll()

	if err != nil {
		response := dtos.ErrorDto{
			Message: "Failed to get shortcuts",
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	response := make([]dtos.ShortcutDto, len(shortcuts))

	for i, shortcut := range shortcuts {
		response[i] = dtos.ShortcutDto{
			Id:    shortcut.Id(),
			Url:   shortcut.Url(),
			Short: h.getShortUrl(shortcut.Id()),
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary	Create Shortcut
// @Tags Shortcut
// @Accept json
// @Produce json
// @Param payload body dtos.CreateShortcutDto true "Create Shortcut"
// @Success 200 {object} dtos.ShortcutDto
// @Failure 400 {object} dtos.ErrorDto
// @Failure 500 {object} dtos.ErrorDto
// @Router /api/shortcut [post]
func (h *ShortcutHandler) Save(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")
	var payload dtos.CreateShortcutDto

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := dtos.ErrorDto{
			Message: "Invalid request payload",
			Error:   err.Error(),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	shortcut, saveError := h.repository.SaveUrl(payload.Url)

	if saveError != nil {
		http.Error(w, saveError.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(dtos.ShortcutDto{
		Id:    shortcut.Id(),
		Url:   shortcut.Url(),
		Short: h.getShortUrl(shortcut.Id()),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary	Get Shortcut by ID
// @Tags Shortcut
// @Produce json
// @Param id path string true "Shortcut ID"
// @Success 200 {object} dtos.ShortcutDto
// @Failure 404 {object} dtos.ErrorDto
// @Failure 500 {object} dtos.ErrorDto
// @Router /api/shortcut/{id} [get]
func (h *ShortcutHandler) FindById(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(r)["id"]

	if shortcut, ok := h.repository.FindById(id); ok {
		if err := json.NewEncoder(w).Encode(dtos.ShortcutDto{
			Id:    shortcut.Id(),
			Url:   shortcut.Url(),
			Short: h.getShortUrl(shortcut.Id()),
		}); err != nil {
			response := dtos.ErrorDto{
				Message: "Failed to get shortcut",
				Error:   err.Error(),
			}

			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	} else {
		response := dtos.ErrorDto{
			Message: "Shortcut not found",
			Error:   "Not Found",
		}
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// @Summary	Delete Shortcut by ID
// @Tags Shortcut
// @Param id path string true "Shortcut ID"
// @Success 204
// @Failure 404 {object} dtos.ErrorDto
// @Failure 500 {object} dtos.ErrorDto
// @Router /api/shortcut/{id} [delete]
func (h *ShortcutHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]

	if ok := h.repository.Delete(id); ok {
		w.WriteHeader(http.StatusNoContent)
	} else {
		response := dtos.ErrorDto{
			Message: "Shortcut not found",
			Error:   "Not Found",
		}
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *ShortcutHandler) getShortUrl(id string) string {
	return h.environment.GetBaseUrl() + "/" + id
}
