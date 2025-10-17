package meta

import (
	"github.com/pkg/errors"
)

type ListOption struct {
	Limit  int `json:"limit"  form:"limit"`
	Offset int `json:"offset" form:"offset"`
}
type ListOptionFn func(opt *ListOption)

func Limit(n int) ListOptionFn {
	return func(opt *ListOption) {
		opt.Limit = n
	}
}

// Page start with 1
func Page(page int, pageSize int) ListOptionFn {
	return func(opt *ListOption) {
		if pageSize > 0 {
			opt.Limit = pageSize
		}
		if page < 1 {
			page = 1
		}
		opt.Offset = (page - 1) * opt.Limit
	}
}

type List[v any] struct {
	Items      []v `json:"items"`
	ListOption `json:",inline"`
	Total      int64 `json:"total,omitempty"`
}

func (l List[v]) Page() int {
	return (l.Offset / l.Limit) + 1
}
func (l List[v]) PageSize() int {
	return l.Limit
}

func (l List[v]) Paged() PagedList[v] {
	return PagedList[v]{
		Items:    l.Items,
		Page:     l.Page(),
		PageSize: l.PageSize(),
	}
}

type PagedList[v any] struct {
	Items    []v `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

var (
	ErrNotImplemented = errors.Errorf("not Implemented")
)
