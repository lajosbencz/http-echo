package httpecho_test

import (
	"testing"

	httpecho "github.com/lajosbencz/http-echo"
)

func TestParseJwtToken(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiJ9.eyJuYW1lIjoiSm9lIENvZGVyIn0.5dlp7GmziL2QS06sZgK4mtaqv0_xX4oFUuTDh1zHK4U"
	jwt, err := httpecho.ParseJwtString(token)
	if err != nil {
		t.Error(err)
	}
	if jwt.Header["alg"] != "HS256" {
		t.Error("header")
	}
	if jwt.Payload["name"] != "Joe Coder" {
		t.Error("payload")
	}
	if jwt.Signature != "5dlp7GmziL2QS06sZgK4mtaqv0_xX4oFUuTDh1zHK4U" {
		t.Error("signature")
	}
}
