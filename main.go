package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/collinewait/ika-gmail-scraper/router"
)

func main() {
	fmt.Println(addNumbers(3, 5))
	r := router.NewRouter()
	log.Fatal(http.ListenAndServe(GetPort(), r))
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

func addNumbers(a, b int) int {
	return a + b
}
