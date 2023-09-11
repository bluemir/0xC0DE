package auth

import (
	"github.com/pkg/errors"
)

var (
	ErrUnauthorized = errors.Errorf("unauthorized")
	ErrForbidden    = errors.Errorf("forbidden")
)
