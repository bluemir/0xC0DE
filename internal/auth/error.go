package auth

import (
	"github.com/pkg/errors"
)

var (
	ErrUnauthroized = errors.Errorf("hashed key not matched")
)
