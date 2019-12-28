package router

import (
	"net/http"

	"github.com/collinewait/ika-gmail-scraper/oauth"
	"github.com/collinewait/ika-gmail-scraper/scraper"
	"github.com/gorilla/mux"
)

type Oauth interface {
	GoogleLogin(w http.ResponseWriter, r *http.Request)
	GoogleCallback(w http.ResponseWriter, r *http.Request)
}

// NewRouter creates new router
func NewRouter() *mux.Router {
	var o Oauth = &oauth.Oauth{}

	r := mux.NewRouter()
	r.HandleFunc("/auth/google/login", o.GoogleLogin)
	r.HandleFunc("/auth/google/callback", o.GoogleCallback)
	r.HandleFunc("/download/attachment", scraper.Scrape)
	return r
}
