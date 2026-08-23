package passwords

import (
	"errors"
	"revit/internal/testingservices"
	"testing"
)

func TestHashingAndVerification(t *testing.T) {
	hashStr, err := MakeHash(testingservices.DEFAULT_PASSWORD)
	if err != nil {
		t.Fatalf("provided a password to be hashed, expected err == nil, found: %v", err)
	}

	_, err = parseArgon2Hash(hashStr)
	if err != nil {
		t.Fatalf("provided a hash to be parsed, expected err == nil, found: %v", err)
	}

	ok, err := VerifyPassword(hashStr, testingservices.DEFAULT_PASSWORD)
	if err != nil {
		t.Fatalf("provided a hash: %s, and password: %s, expected err == nil, found: %v", hashStr, testingservices.DEFAULT_PASSWORD, err)
	}
	if !ok {
		t.Fatalf("provided hash: %s, and password: %s, but found match: %t", hashStr, testingservices.DEFAULT_PASSWORD, ok)
	}
}

func TestHashParsingError(t *testing.T) {
	_, err := parseArgon2Hash("WRONG")
	if err == nil {
		t.Fatalf("provided an invalid hash to be parsed, expected err != nil, found: %v", err)
	}
	expectedErr := &Argon2Error{}
	if ok := errors.As(err, expectedErr); !ok {
		t.Fatalf("provided an invalid hash to be parsed, expected type of the error == %T, found type of err == %T", expectedErr, err)
	} else {
		if expectedErr.errType != Argon2InvalidHash {
			t.Fatalf("provided an invalid hash to be parsed, expected err.errType == %s, found: %s", Argon2InvalidHash.String(), expectedErr.errType)
		}
	}
}
