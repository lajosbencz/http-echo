package httpecho_test

import (
	"testing"

	httpecho "github.com/lajosbencz/http-echo"
)

const jwtToken = "eyJhbGciOiJIUzI1NiJ9.eyJuYW1lIjoiSm9lIENvZGVyIn0.5dlp7GmziL2QS06sZgK4mtaqv0_xX4oFUuTDh1zHK4U"

func TestParseJwtToken(t *testing.T) {
	jwt, err := httpecho.ParseJwtString(jwtToken)
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
