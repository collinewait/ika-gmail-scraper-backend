package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_extractToken_shouldReturnToken(t *testing.T) {

	expectedToken := "sometokenhere"

	var bearer = "Bearer " + expectedToken
	r := httptest.NewRequest(http.MethodGet, "/urlhere", nil)
	r.Header.Add("Authorization", bearer)

	token, _ := extractToken(r)
	if token != expectedToken {
		t.Errorf("extractToken() = %v, want %v", token, expectedToken)
	}
}

func Test_extractToken_shouldReturnError(t *testing.T) {

	expectedError := "Bearer token not in proper format"

	var bearer = "sometokenhere"
	r := httptest.NewRequest(http.MethodGet, "/urlhere", nil)
	r.Header.Add("Authorization", bearer)

	_, err := extractToken(r)
	if err.Error() != expectedError {
		t.Errorf("extractToken() = %v, want %v", err.Error(), expectedError)
	}
}
