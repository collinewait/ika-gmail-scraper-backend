package oauth

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dchest/uniuri"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var (
	googleOauthConfig *oauth2.Config
)
var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func init() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "https://ika-gmail-scraper-backend.herokuapp.com/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{gmail.GmailReadonlyScope},
		Endpoint:     google.Endpoint,
	}

	gob.Register(&time.Time{})

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure: true,
		SameSite: http.SameSiteNoneMode,
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
	url := googleOauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline)
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
	err := oauth.saveoauthTokensInSession(w, r, oauth2Token)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusInternalServerError)
		return
	}

	jwtToken, _ := oauth.generateJwtToken()

	http.Redirect(w, r, os.Getenv("FRONTEND_REDIRECT_URL")+"?access_token="+jwtToken, http.StatusFound)
}

var getToken = func(code string) *oauth2.Token {
	token, err := googleOauthConfig.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token: %v", err)
	}
	return token
}

// GetGmailService will return a gmail service.
func GetGmailService(token, refreshToken string, expiry time.Time) *gmail.Service {
	ctx := context.Background()
	oauth2Token := &oauth2.Token{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
	service, err := gmail.NewService(ctx, option.WithTokenSource(googleOauthConfig.TokenSource(ctx, oauth2Token)))
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
	expirationTime := time.Now().Add(6 * time.Hour)
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

//DecodeJwtToken Parse the JWT string and store the result in claims.
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

func (oauth *Oauth) saveoauthTokensInSession(
	w http.ResponseWriter,
	r *http.Request,
	oauth2Token *oauth2.Token,
) error {

	sessionTkn, err := store.Get(r, "oauth-tkn-session")
	if err != nil {
		fmt.Println("Error creating oauth-tkn-session: ", err.Error())
		return err
	}
	sessionRtkn, err := store.Get(r, "session-rtkn-session")
	if err != nil {
		fmt.Println("Error creating session-rtkn-session: ", err.Error())
		return err
	}
	sessionExp, err := store.Get(r, "expiry-session")
	if err != nil {
		fmt.Println("Error creating expiry-session: ", err.Error())
		return err
	}

	sessionTkn.Values[oauth.stateString+"AccessToken"] = oauth2Token.AccessToken
	sessionRtkn.Values[oauth.stateString+"RefreshToken"] = oauth2Token.RefreshToken
	sessionExp.Values[oauth.stateString+"Expiry"] = oauth2Token.Expiry

	err = sessionTkn.Save(r, w)
	if err != nil {
		fmt.Println("Error saving session-tkn-session: ", err.Error())
		return err
	}

	err = sessionRtkn.Save(r, w)
	if err != nil {
		fmt.Println("Error saving session-rtkn-session: ", err.Error())
		return err
	}
	err = sessionExp.Save(r, w)
	if err != nil {
		fmt.Println("Error saving expiry-session: ", err.Error())
		return err
	}

	return nil
}

// RetrieveTokensFromSession returns a token stored in the session
func RetrieveTokensFromSession(
	r *http.Request,
	randomID string,
) (string, string, time.Time, error) {
	sessionTkn, err := store.Get(r, "oauth-tkn-session")
	if err != nil {
		return "", "", time.Now(), err
	}
	sessionRtkn, err := store.Get(r, "session-rtkn-session")
	if err != nil {
		return "", "", time.Now(), err
	}
	sessionExp, err := store.Get(r, "expiry-session")
	if err != nil {
		return "", "", time.Now(), err
	}

	tkn, ok := sessionTkn.Values[randomID+"AccessToken"].(string)
	if !ok {
		return "", "", time.Now(), errors.New("received unexpected type of token")
	}

	rtkn, ok := sessionRtkn.Values[randomID+"RefreshToken"].(string)
	if !ok {
		return "", "", time.Now(), errors.New("received unexpected type of refresh token")
	}

	expiry, ok := sessionExp.Values[randomID+"Expiry"].(*time.Time)
	if !ok {
		return "", "", time.Now(), errors.New("received unexpected type of token expiry time")
	}

	return tkn, rtkn, *expiry, nil
}
