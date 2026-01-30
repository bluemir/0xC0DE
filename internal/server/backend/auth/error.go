package auth

import (
	"github.com/cockroachdb/errors"
)

var (
	ErrUnauthorized = errors.Errorf("unauthorized")
	ErrForbidden    = errors.Errorf("forbidden")
)
