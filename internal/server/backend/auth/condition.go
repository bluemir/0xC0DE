package auth

import (
	"github.com/expr-lang/expr"
	"github.com/pkg/errors"
)

type Condition string

func (cond Condition) IsMatched(ctx Context) (bool, error) {
	p, err := expr.Compile(string(cond), expr.AsBool())
	if err != nil {
		return false, errors.WithStack(err)
	}
	result, err := expr.Run(p, ctx)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return result.(bool), nil
}
