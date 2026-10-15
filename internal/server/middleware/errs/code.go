package errs

import (
	"net/http"
	"os"
	"strings"

	// 저장소 전체가 cockroachdb/errors 로 감싸 올린다. 표준 errors.Is 는
	// 그쪽 마커(errors.Mark)를 읽지 못해, 표시해 둔 에러가 여기서 500 으로 떨어진다.
	"github.com/cockroachdb/errors"

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
