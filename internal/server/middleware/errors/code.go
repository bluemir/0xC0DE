package errors

import (
	"net/http"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/auth"
)

func findCode(err error) int {
	switch {
	case errors.Is(err, auth.ErrUnauthroized):
		return http.StatusUnauthorized
	case errors.Is(err, gorm.ErrRecordNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
