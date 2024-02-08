package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/luiscib3r/shortly/app/docs"
	"github.com/luiscib3r/shortly/app/internal/data/datasources"
	"github.com/luiscib3r/shortly/app/internal/data/repositories"
	"github.com/luiscib3r/shortly/app/internal/domain/entities"
	"github.com/luiscib3r/shortly/app/internal/presentation/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Swagger
//	@title			Shortly Service
//	@version		1.0
//	@description	URL shortener service

//	@contact.name	Luis Ciber
//	@contact.url	https://www.luisciber.com/
//	@contact.email	luisciber640@gmail.com

//	@license.name	MIT
//	@license.url	https://github.com/luicib3r/shortly

// @host
// @BasePath	/
func main() {

	//----------------------------------------
	// Datasources
	//----------------------------------------
	shortcutMemDB := datasources.NewMemDB[entities.Shortcut]()
	environment := datasources.NewEnvironmentDataSource()
	shortcutDynamoDB, err := datasources.NewShortcutDynamoDB()

	if err != nil {
		log.Fatal(err)
	}

	//----------------------------------------
	// Repositories
	//----------------------------------------
	shortcutRepository := repositories.NewShortcutRepositoryData(shortcutDynamoDB, shortcutMemDB)
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
	// GET /api/shortcut
	//----------------------------------------
	router.Methods("GET").Path("/api/shortcut").HandlerFunc(shortcutHandler.FindAll)
	//----------------------------------------
	// POST /api/shortcut
	//----------------------------------------
	router.Methods("POST").Path("/api/shortcut").HandlerFunc(shortcutHandler.Save)
	//----------------------------------------
	// GET /api/shortcut/{id}
	//----------------------------------------
	router.Methods("GET").Path("/api/shortcut/{id:[a-zA-Z0-9]{6}}").HandlerFunc(shortcutHandler.FindById)
	//----------------------------------------
	// DELETE /api/shortcut/{id}
	//----------------------------------------
	router.Methods("DELETE").Path("/api/shortcut/{id:[a-zA-Z0-9]{6}}").HandlerFunc(shortcutHandler.Delete)

	//----------------------------------------
	// Redirect
	redirectHandler := handlers.NewRedirectHandler(
		shortcutRepository,
	)
	//----------------------------------------
	// GET /{id}
	//----------------------------------------
	router.Methods("GET").Path("/{id:[a-zA-Z0-9]{6}}").HandlerFunc(redirectHandler.Redirect)

	// Swagger
	router.Methods("GET").PathPrefix("/docs").Handler(httpSwagger.WrapHandler)

	// Health
	router.Methods("GET").Path("/healthcheck").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

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
