package util

import (
	"crypto"
	"encoding/hex"
	"io"
)

func Hash(str string) string {
	hashed := crypto.SHA512.New()
	io.WriteString(hashed, str)
	return hex.EncodeToString(hashed.Sum(nil))

}
