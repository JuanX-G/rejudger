package passwords

import (
	"errors"
	"revit/internal/testingservices"
	"strings"
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

func TestVerificationError(t *testing.T) {
	_, err := VerifyPassword("WRONG", "NOT_HERE_EITHER")
	if err == nil {
		t.Fatalf("provided an invalid hash, for a password, to be verified against, expected err != nil, found: %v", err)
	}
	expectedErr := &Argon2Error{}
	if ok := errors.As(err, expectedErr); !ok {
		t.Fatalf("provided an invalid hash to be verified against, expected type of the error == %T, found type of err == %T", expectedErr, err)
	} else {
		if expectedErr.errType != Argon2InvalidHash {
			t.Fatalf("provided an invalid hash to be used for verification, expected err.errType == %s, found: %s", Argon2InvalidHash.String(), expectedErr.errType)
		}
	}
}

func TestParsingInvalidVersion(t *testing.T) {
	hash, err := MakeHash(testingservices.DEFAULT_PASSWORD)
	if err != nil {
		t.Fatalf("provided a password to be hashed, expected err == nil, found: %v", err)
	}

	// create a hash with invalid algorithm
	parts := strings.Split(hash, "$")
	parts[1] = "no-argon-here"
	hash2 := strings.Join(parts, "$")

	_, err = VerifyPassword(hash2, "NOT_HERE_EITHER") // test with the invalid algorithm string
	if err == nil {
		t.Fatalf("provided an invalid hash, for a password, to be verified against, expected err != nil, found: %v", err)
	}

	expectedErr := &Argon2Error{}
	if ok := errors.As(err, expectedErr); !ok {
		t.Fatalf("provided a hash with invalid algorithm to be verifiend against, expected type of the error == %T, found type of err == %T", expectedErr, err)
	} else {
		if expectedErr.errType != Argon2UnsupportedAlgorithm {
			t.Fatalf("provided a hash with invalid algorithm hash to be verified against, expected err.errType == %s, found: %s", Argon2UnsupportedAlgorithm.String(), expectedErr.errType)
		}
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
