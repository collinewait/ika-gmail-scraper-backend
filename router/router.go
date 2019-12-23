package router

import (
	"github.com/collinewait/ika-gmail-scraper/oauth"
	"github.com/gorilla/mux"
)

// NewRouter creates new router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/auth/google/login", oauth.GoogleLogin)
	r.HandleFunc("/auth/google/callback", oauth.GoogleCallback)
	return r
}
