package auth

import (
	"github.com/pkg/errors"
)

var (
	ErrKeyNotMatched = errors.Errorf("hashed key not matched")
)
