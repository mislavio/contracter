package helpers

import (
	"math/rand"
	"time"
)

// Adapted from https://www.calhoun.io/creating-random-strings-in-go/

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// stringWithCharset generates random string of given length from character set
func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// RandomString generates random string of given length
func RandomString(length int) string {
	return stringWithCharset(length, charset)
}
