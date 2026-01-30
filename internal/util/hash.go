package util

import (
	"crypto"
	"encoding/hex"
	"io"
)

func Hash(strs ...string) string {
	hashed := crypto.SHA512.New()
	for _, str := range strs {
		_, _ = io.WriteString(hashed, str)
	}
	return hex.EncodeToString(hashed.Sum(nil))

}
