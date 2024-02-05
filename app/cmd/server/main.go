package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/luiscib3r/shortly/app/handlers"
)

func main() {
	mux := http.ServeMux{}

	mux.Handle(handlers.RootPath, &handlers.RootHandler{})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port " + port)
	if err := http.ListenAndServe(":"+port, &mux); err != nil {
		panic(err)
	}

}
