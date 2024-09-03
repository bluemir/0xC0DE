package handler

import (
	"github.com/bluemir/0xC0DE/internal/server/injector"
)

var backends = injector.Backends

type ListResponse[T any] struct {
	Items []T
	//Page int
	//PageSize int
}
