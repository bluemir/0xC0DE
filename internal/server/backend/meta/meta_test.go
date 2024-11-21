package meta_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

func TestPageOption(t *testing.T) {
	opts := []meta.ListOptionFn{
		meta.Page(1, 200),
	}

	opt := meta.ListOption{}
	for _, fn := range opts {
		fn(&opt)
	}

	assert.Equal(t, 200, opt.Limit)
	assert.Equal(t, 0, opt.Offset)
}
func TestPageOptionWithZero(t *testing.T) {
	opts := []meta.ListOptionFn{
		meta.Page(0, 200), // equal to 1 page.
	}

	opt := meta.ListOption{}
	for _, fn := range opts {
		fn(&opt)
	}

	assert.Equal(t, 200, opt.Limit)
	assert.Equal(t, 0, opt.Offset)
}
func TestPageOptionWith10Page(t *testing.T) {
	// page01: 00~09
	// page02: 10~19
	// ...
	// page10: 90~99
	opts := []meta.ListOptionFn{
		meta.Page(10, 10),
	}

	opt := meta.ListOption{}
	for _, fn := range opts {
		fn(&opt)
	}

	assert.Equal(t, 10, opt.Limit)
	assert.Equal(t, 90, opt.Offset)
}
func TestListOptionToPage(t *testing.T) {
	opt := meta.ListOption{
		Limit:  100,
		Offset: 0,
	}
	list := meta.List[any]{
		ListOption: opt,
	}

	assert.Equal(t, 1, list.Page())
	assert.Equal(t, 100, list.PageSize())
}

func TestListOptionToPageWithMultiplePage(t *testing.T) {
	type Result struct {
		Page     int
		PageSize int
	}

	testcase := []struct {
		Option meta.ListOption
		Result Result
	}{
		{
			// see TestPageOptionWith10Page
			Option: meta.ListOption{
				Limit:  10,
				Offset: 90,
			},
			Result: Result{
				Page:     10,
				PageSize: 10,
			},
		},
		{
			//wired case, but just ceil number
			Option: meta.ListOption{
				Limit:  100,
				Offset: 13,
			},
			Result: Result{
				Page:     1,
				PageSize: 100,
			},
		},
	}

	for _, c := range testcase {
		list := meta.List[any]{
			ListOption: c.Option,
		}

		assert.Equal(t, c.Result.Page, list.Page())
		assert.Equal(t, c.Result.PageSize, list.PageSize())
	}
}
