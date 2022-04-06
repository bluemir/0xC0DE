package errors

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestT(t *testing.T) {
	re := errors.Wrap(errors.New("example"), "goo")

	e := &gin.Error{Err: errors.WithStack(re)}

	result := getStackTrace(e)
	assert.Len(t, result, 3)
}
