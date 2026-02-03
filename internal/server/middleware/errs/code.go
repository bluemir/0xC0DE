package errs

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/go-playground/validator/v10"
	"github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// code returns the HTTP status code for any error.
func code(err error) int {
	var (
		sqliteErr      sqlite3.Error
		validationErrs validator.ValidationErrors
	)

	switch {
	// errors.Is: sentinel errors (exact match)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return http.StatusConflict
	case errors.Is(err, auth.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, auth.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, os.ErrNotExist):
		return http.StatusNotFound
	case errors.Is(err, meta.ErrNotImplemented):
		return http.StatusNotImplemented

	// errors.As: type-based errors
	case errors.As(err, &validationErrs):
		return http.StatusBadRequest
	case errors.As(err, &sqliteErr):
		switch sqliteErr.Code {
		case sqlite3.ErrConstraint:
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return http.StatusConflict
			}
			return http.StatusBadRequest
		}
		return http.StatusNotImplemented
	}

	// finally check string match
	logrus.Trace(err.Error())
	switch {
	case strings.HasPrefix(err.Error(), "html/template: ") && strings.HasSuffix(err.Error(), " is undefined"):
		//html/template: ".*" is undefined
		return http.StatusNotImplemented
	}

	return http.StatusInternalServerError
}
