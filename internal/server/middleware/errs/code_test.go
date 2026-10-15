package errs

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

// code 가 에러 종류마다 어떤 상태를 내는지 고정한다.
// 감싼 것도 같이 본다. 저장소 전체가 errors.WithStack 을 거쳐 올라오기 때문이다.
func TestCode(t *testing.T) {
	validationErr := validator.New().Struct(struct {
		Name string `validate:"required"`
	}{})

	for name, tc := range map[string]struct {
		err  error
		want int
	}{
		"duplicated key":      {gorm.ErrDuplicatedKey, http.StatusConflict},
		"unauthorized":        {auth.ErrUnauthorized, http.StatusUnauthorized},
		"forbidden":           {auth.ErrForbidden, http.StatusForbidden},
		"record not found":    {gorm.ErrRecordNotFound, http.StatusNotFound},
		"file not exist":      {os.ErrNotExist, http.StatusNotFound},
		"not implemented":     {meta.ErrNotImplemented, http.StatusNotImplemented},
		"validation":          {validationErr, http.StatusBadRequest},
		"unique constraint":   {sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}, http.StatusConflict},
		"other constraint":    {sqlite3.Error{Code: sqlite3.ErrConstraint}, http.StatusBadRequest},
		"undefined template":  {errors.Newf(`html/template: "foo" is undefined`), http.StatusNotImplemented},
		"unknown":             {errors.New("something else"), http.StatusInternalServerError},
		"wrapped auth":        {errors.WithStack(auth.ErrUnauthorized), http.StatusUnauthorized},
		"wrapped not found":   {errors.Wrap(gorm.ErrRecordNotFound, "while loading"), http.StatusNotFound},
		"wrapped constraint":  {errors.WithStack(sqlite3.Error{Code: sqlite3.ErrConstraint, ExtendedCode: sqlite3.ErrConstraintUnique}), http.StatusConflict},
		"joined with auth":    {errors.Join(auth.ErrUnauthorized, errors.New("cause")), http.StatusUnauthorized},
		"secondary with auth": {errors.WithSecondaryError(auth.ErrUnauthorized, errors.New("cause")), http.StatusUnauthorized},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, code(tc.err))
		})
	}
}

// 마커로 표시한 에러도 그 상태로 나가야 한다.
//
// auth 는 errors.Mark 를 쓰지 않는다(ADR-0009). 그래도 이 테스트를 둔다.
// code 가 표준 errors 로 돌아가면 마커가 묻히고, 그것이 전에 passkey 실패를
// 500 으로 내보내던 원인이었다. 여기서 막는다.
func TestCodeReadsMarker(t *testing.T) {
	marked := errors.Mark(errors.New("crypto/bcrypt: hashedPassword is not the hash"), auth.ErrUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, code(marked))
}

// nil 은 여기까지 오지 않지만, 와도 터지지는 않아야 한다
func TestCodeNil(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = code(errors.New(fmt.Sprint(nil)))
	})
}
