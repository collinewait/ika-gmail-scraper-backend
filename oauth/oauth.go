package oauth

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/dchest/uniuri"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var (
	googleOauthConfig *oauth2.Config
)

func init() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "https://ika-gmail-scraper-backend.herokuapp.com/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{gmail.GmailReadonlyScope},
		Endpoint:     google.Endpoint,
	}
}

type Oauth struct {
	stateString string
}

func (oauth *Oauth) generateRandomString() string {
	s := uniuri.New()
	return s
}

func (oauth *Oauth) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthStateString := oauth.generateRandomString()
	oauth.stateString = oauthStateString
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (oauth *Oauth) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != oauth.stateString {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token := getToken(code)

	http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"?access_token="+token.AccessToken, http.StatusFound)
}

var getToken = func(code string) *oauth2.Token {
	token, err := googleOauthConfig.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token: %v", err)
	}
	return token
}

// GetGmailService will return a gmail service.
func GetGmailService(token string) *gmail.Service {
	ctx := context.Background()
	oauthToken := &oauth2.Token{
		AccessToken: token,
	}
	service, err := gmail.NewService(ctx, option.WithTokenSource(googleOauthConfig.TokenSource(ctx, oauthToken)))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
	return service
}
