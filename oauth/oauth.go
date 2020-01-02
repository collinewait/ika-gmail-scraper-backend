package oauth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dchest/uniuri"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var (
	googleOauthConfig *oauth2.Config
)
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

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
	oauth2Token := getToken(code)

	jwtToken, _ := oauth.generateJwtToken()

	fmt.Printf("jwtToken: %s. oauth2Token: %s", jwtToken, oauth2Token)

	http.Redirect(w, r, os.Getenv("FRONTEND_URL")+"?access_token="+jwtToken, http.StatusFound)
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

// Claims struct to be encoded to a JWT.
type Claims struct {
	RandomID string
	jwt.StandardClaims
}

func (oauth *Oauth) generateJwtToken() (string, error) {
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &Claims{
		RandomID: oauth.stateString,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//DecodeJwtToken Parse the JWT string and store the result in `claims`.
func DecodeJwtToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, errors.New(err.Error())
	}
	if !tkn.Valid {
		return nil, errors.New(err.Error())
	}

	return claims, nil
}
