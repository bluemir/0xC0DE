package util

import (
	"crypto/rand"
	"math/big"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandomString generates a cryptographically secure random alphanumeric string of the given length.
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	letterLength := big.NewInt(int64(len(letterBytes)))

	for index := 0; index < length; index++ {
		randomNum, err := rand.Int(rand.Reader, letterLength)
		if err != nil {
			panic(err)
		}
		result[index] = letterBytes[randomNum.Int64()]
	}

	return string(result)
}
