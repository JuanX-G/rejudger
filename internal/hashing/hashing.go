package hashing

import (
	"encoding/base64"
	"time"

	"golang.org/x/crypto/sha3"
)

func HashSubmission(text string, authorName string) string {
	t := time.Now().Format(time.RFC1123)
	hash := sha3.Sum512([]byte(text + authorName + t))
	return base64.StdEncoding.EncodeToString(hash[:])
}
