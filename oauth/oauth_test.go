package oauth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func Test_GoogleCallback_shouldReturnAString(t *testing.T) {

	r := httptest.NewRequest(http.MethodGet, "/auth/google/callback?state=pseudo-random&code=somecodehere", nil)
	w := httptest.NewRecorder()

	expected := "Can call the api now"

	o := &Oauth{
		stateString: "pseudo-random",
	}

	getClient = func(config *oauth2.Config, code string) *http.Client {
		return &http.Client{}
	}

	o.GoogleCallback(w, r)

	fmt.Println(w.Result().StatusCode)
	resBody, _ := ioutil.ReadAll(w.Result().Body)

	if string(resBody) != expected {
		t.Errorf("GoogleCallback() = %v, want %v", resBody, expected)
	}
}
