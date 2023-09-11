package auth

import (
	"github.com/antonmedv/expr"
)

type Condition string

func (cond Condition) IsMatched(ctx Context) (bool, error) {
	p, err := expr.Compile(string(cond), expr.AsBool())
	if err != nil {
		return false, err
	}
	result, err := expr.Run(p, ctx)
	if err != nil {
		return false, err
	}
	return result.(bool), nil
}
