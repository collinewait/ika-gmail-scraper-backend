package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/collinewait/ika-gmail-scraper/router"
	"github.com/gorilla/handlers"
)

func main() {
	r := router.NewRouter()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Origin"})
	methods := handlers.AllowedMethods([]string{"GET", "POST"})
	origins := handlers.AllowedOrigins([]string{"https://accounts.google.com", os.Getenv("FRONTEND_URL")})
	allowCreds := handlers.AllowCredentials()
	log.Fatal(http.ListenAndServe(GetPort(), handlers.CORS(headers, methods, origins, allowCreds)(r)))
}

// GetPort gets the Port from the environment so we can run on Heroku
func GetPort() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "4747"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}
