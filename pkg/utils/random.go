package utils

import (
	"math/rand"
	"time"
)

const (
	// The set of characters to use for generating random strings.
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// The length of the generated alias.
	aliasLength = 6
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// String generates a random string of a fixed length.
func String() string {
	b := make([]byte, aliasLength)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}