package passwords

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	Argon2Time      = 1
	Argon2Memory    = 128 * 1024
	Argon2Threads   = 2
	Argon2KeyLength = 32
)

// Parsed argon2id hash
type Argon2Hash struct {
	Hash                 []byte // Actual hash
	Salt                 []byte // Salt used for the hash
	Time, Memory, KeyLen uint32 // Argon2id params
	Threads              uint8  // Ardon2id param
}

// returns a fomatted hash that hold data neccessary for
// proper password verification later.
func MakeHash(pass string) (string, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(pass), salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		Argon2Memory,
		Argon2Time,
		Argon2Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encodedHash, nil
}

//go:generate stringer -type=Argon2ErrorType
type Argon2ErrorType int

const (
	Argon2InvalidHash Argon2ErrorType = iota
	Argon2UnsupportedAlgorithm
)

type Argon2Error struct {
	errType Argon2ErrorType
}

func (a Argon2Error) Error() string {
	switch a.errType {
	case Argon2InvalidHash:
		return "invalid hash"
	case Argon2UnsupportedAlgorithm:
		return "unsupported algorithm"
	default:
		return "unknown error"
	}
}

// Parse a hash string from MakeHash into a struct.
func parseArgon2Hash(encodedHash string) (Argon2Hash, error) {
	components := strings.Split(encodedHash, "$")
	if len(components) != 6 {
		return Argon2Hash{}, Argon2Error{errType: Argon2InvalidHash}
	}

	if !strings.HasPrefix(components[1], "argon2id") {
		return Argon2Hash{}, Argon2Error{errType: Argon2UnsupportedAlgorithm}
	}

	var version int
	fmt.Sscanf(components[2], "v=%d", &version)

	config := Argon2Hash{}
	fmt.Sscanf(components[3], "m=%d,t=%d,p=%d",
		&config.Memory, &config.Time, &config.Threads)

	salt, err := base64.RawStdEncoding.DecodeString(components[4])
	if err != nil {
		return Argon2Hash{}, fmt.Errorf("salt decoding failed: %w", err)
	}
	config.Salt = salt

	hash, err := base64.RawStdEncoding.DecodeString(components[5])
	if err != nil {
		return Argon2Hash{}, fmt.Errorf("hash decoding failed: %w", err)
	}
	config.Hash = hash
	config.KeyLen = uint32(len(hash))

	return config, nil
}

// Check if the pass corresponds to the storedHash. Stored hash should be of
// the format given my MakeHash.
func VerifyPassword(storedHash, pass string) (bool, error) {
	config, err := parseArgon2Hash(storedHash)
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey(
		[]byte(pass),
		config.Salt,
		config.Time,
		config.Memory,
		config.Threads,
		config.KeyLen,
	)

	match := subtle.ConstantTimeCompare(config.Hash, computedHash) == 1
	return match, nil
}
