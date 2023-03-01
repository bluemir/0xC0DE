package auth

import (
	"github.com/pkg/errors"
)

var (
	ErrUnauthroized = errors.Errorf("hashed key not matched")
	ErrNotFound     = errors.Errorf("not found")
	ErrNotAllowed   = errors.Errorf("not allowed method")
)
