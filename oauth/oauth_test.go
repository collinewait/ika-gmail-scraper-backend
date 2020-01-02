package oauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func Test_GoogleLogin_shouldRedirect(t *testing.T) {

	r := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	w := httptest.NewRecorder()

	expectedStatusCode := 307

	o := &Oauth{}
	o.GoogleLogin(w, r)

	actualStatusCode := w.Result().StatusCode

	if actualStatusCode != expectedStatusCode {
		t.Errorf("GoogleLogin() = %v, want %v", actualStatusCode, expectedStatusCode)
	}
}

func Test_GoogleCallback_shouldRedirectOnWrongState(t *testing.T) {

	r := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=wrong_state", nil)
	w := httptest.NewRecorder()

	expectedStatusCode := 307

	o := &Oauth{}
	o.GoogleCallback(w, r)

	actualStatusCode := w.Result().StatusCode

	if actualStatusCode != expectedStatusCode {
		t.Errorf("GoogleCallback() = %v, want %v", actualStatusCode, expectedStatusCode)
	}
}

func Test_GoogleCallback_shouldRedirect(t *testing.T) {

	r := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=pseudo-random&code=somecodehere", nil)
	w := httptest.NewRecorder()

	expectedStatusCode := 302

	o := &Oauth{
		stateString: "pseudo-random",
	}

	getToken = func(code string) *oauth2.Token {
		return &oauth2.Token{}
	}

	o.GoogleCallback(w, r)

	actualStatusCode := w.Result().StatusCode

	if actualStatusCode != expectedStatusCode {
		t.Errorf("GoogleCallback() = %v, want %v", actualStatusCode, expectedStatusCode)
	}
}

func Test_GetGmailService(t *testing.T) {
	expectedBasePathValue := "https://www.googleapis.com/gmail/v1/users/"
	actualValue := GetGmailService("sometokenhere")
	if actualValue.BasePath != expectedBasePathValue {
		t.Errorf("GetGmailService() = %v, want %v", actualValue, expectedBasePathValue)
	}
}

func Test_generateJwtToken_shouldReturnToken(t *testing.T) {
	o := &Oauth{
		stateString: "pseudo-random",
	}

	jwtToken, _ := o.generateJwtToken()

	if len(strings.Split(jwtToken, ".")) != 3 {
		t.Errorf("generateJwtToken() expected some value but got an empty string")
	}
}

func Test_DecodeJwtToken_shouldReturnAClaim(t *testing.T) {
	o := &Oauth{
		stateString: "pseudo-random",
	}

	jwtToken, _ := o.generateJwtToken()

	claim, _ := DecodeJwtToken(jwtToken)

	if claim.RandomID != o.stateString {
		t.Errorf("DecodeJwtToken() = %v, want %v", claim.RandomID, o.stateString)
	}
}

func Test_DecodeJwtToken_shouldReturnAnError(t *testing.T) {
	_, err := DecodeJwtToken("invalidJwtToken")

	if err == nil {
		t.Errorf("DecodeJwtToken() expected an error but it wasn't returned")
	}
}
