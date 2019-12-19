package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	sum := addNumbers(2, 6)
	fmt.Println(sum)

	http.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(GetPort(), nil))
}

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world") // nolint
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
