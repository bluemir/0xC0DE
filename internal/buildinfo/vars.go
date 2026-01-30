package buildinfo

import (
	"crypto"
	"encoding/hex"
	"io"
)

var (
	Version   string
	AppName   string
	BuildTime string
)

func Signature() string {
	hashed := crypto.SHA512.New()

	_, _ = io.WriteString(hashed, AppName)
	_, _ = io.WriteString(hashed, Version)
	_, _ = io.WriteString(hashed, BuildTime)

	return hex.EncodeToString(hashed.Sum(nil))
}
