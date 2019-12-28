package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Scrape will extract attachments contained in mails sent by a specific email.
func Scrape(w http.ResponseWriter, r *http.Request) {
	token, err := extractToken(r)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token)) // nolint
}

func extractToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		return "", errors.New("Bearer token not in proper format")
	}

	reqToken = strings.TrimSpace(splitToken[1])
	return reqToken, nil
}
