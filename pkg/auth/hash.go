package auth

import (
	"crypto"
	"encoding/hex"
	"io"
	"strings"
)

func hashRawHex(str string) string {
	hashed := crypto.SHA512.New()
	io.WriteString(hashed, str)
	return hex.EncodeToString(hashed.Sum(nil))
}
func hash(str string, salt ...string) string {
	return hashRawHex(strings.Join(append([]string{str}, salt...), "/"))
}
