package main

import (
	"strconv"
	"testing"
)

func Test_Authenticate(t *testing.T) {

	user := &AuthUser{
		Username: "user",
		Password: "pass",
		valid:    false,
	}
	user.Authenticate()
	if user.valid {
		t.Fatalf("auth error %s, expected is false", strconv.FormatBool(user.valid))
	}
	user.Password = "www"
	user.Authenticate()
	if !user.valid {
		t.Fatalf("auth error %s, expected is true", strconv.FormatBool(user.valid))
	}
}
