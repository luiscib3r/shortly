package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/luiscib3r/shortly/app/internal/data/datasources"
	"github.com/luiscib3r/shortly/app/internal/data/repositories"
	"github.com/luiscib3r/shortly/app/internal/domain/entities"
	"github.com/luiscib3r/shortly/app/internal/presentation/handlers"
)

func main() {
	//----------------------------------------
	// Datasources
	//----------------------------------------
	shortcutMemDB := datasources.NewMemDB[entities.Shortcut]()
	environment := datasources.NewEnvironmentDataSource()

	//----------------------------------------
	// Repositories
	//----------------------------------------
	shortcutRepository := repositories.NewShortcutRepositoryData(shortcutMemDB)
	environmentRepository := repositories.NewEnvironmentRepositoryData(environment)

	//----------------------------------------
	// Router
	//----------------------------------------
	router := mux.NewRouter()

	//----------------------------------------
	// Routes
	//----------------------------------------
	// GET /
	rootHandler := &handlers.RootHandler{}
	router.Methods("GET").Path("/").Handler(rootHandler)
	//----------------------------------------
	// Shortcut API
	//----------------------------------------
	shortcutHandler := handlers.NewShortcutHandler(
		shortcutRepository,
		environmentRepository,
	)
	//----------------------------------------
	// GET /api/shortcuts
	//----------------------------------------
	router.Methods("GET").Path("/api/shortcut").HandlerFunc(shortcutHandler.FindAll)
	//----------------------------------------
	// POST /api/shortcuts
	//----------------------------------------
	router.Methods("POST").Path("/api/shortcut").HandlerFunc(shortcutHandler.Save)
	//----------------------------------------
	// GET /api/shortcut/{id}
	//----------------------------------------
	router.Methods("GET").Path("/api/shortcut/{id}").HandlerFunc(shortcutHandler.FindById)
	//----------------------------------------
	// DELETE /api/shortcut/{id}
	//----------------------------------------
	router.Methods("DELETE").Path("/api/shortcut/{id}").HandlerFunc(shortcutHandler.Delete)

	//----------------------------------------
	// Redirect
	redirectHandler := handlers.NewRedirectHandler(
		shortcutRepository,
	)
	//----------------------------------------
	// GET /{id}
	//----------------------------------------
	router.Methods("GET").Path("/{id}").HandlerFunc(redirectHandler.Redirect)

	//----------------------------------------
	// Server
	//----------------------------------------
	port := environment.GetEnvironment().PORT
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Println("Server running on ", addr)
	log.Fatal(srv.ListenAndServe())
}
