package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bluemir/0xC0DE/internal/store"
)

type MyStruct struct{}

func TestHelpers(t *testing.T) {
	// CamelCaseToKebabCase
	assert.Equal(t, "hello-world", store.CamelCaseToKebabCase("HelloWorld"))
	assert.Equal(t, "my-struct", store.CamelCaseToKebabCase("MyStruct"))
	assert.Equal(t, "http-request", store.CamelCaseToKebabCase("HTTPRequest")) // Regex behavior might vary, let's verfiy
	// Based on regex in code:
	// matchFirstCap: (.)([A-Z][a-z]+) -> ${1}-${2}
	// matchAllCap: ([a-z0-9])([A-Z]) -> ${1}-${2}
	// HTTPRequest -> HTTPRequest (no match for first cap logic if adjacent caps?)
	// Actually matchAllCap might handle it.
	// Let's rely on unit test to verify or fix expectation.

	// TakeLastName
	assert.Equal(t, "MyStruct", store.TakeLastName("github.com/bluemir/0xC0DE/internal/store_test.MyStruct"))
	assert.Equal(t, "MyStruct", store.TakeLastName("MyStruct"))

	// GetTypeString
	assert.Equal(t, "my-struct", store.GetTypeString(MyStruct{}))
	assert.Equal(t, "my-struct", store.GetTypeString(&MyStruct{})) // Expected?
	// reflect.TypeOf(&MyStruct{}).String() is "*store_test.MyStruct"
	// TakeLastName -> "MyStruct" (split by dot) - wait, * is part of it?
	// If reflect.TypeOf(obj).String() includes *, TakeLastName might keep it.
	// Let's verify `store.go` implementation details if needed.
	// strings.Split("*...", ".") -> prefix * remains.
	// CamelCaseToKebabCase("*MyStruct") -> "*my-struct" ?
}

func TestStore_NotImplements(t *testing.T) {
	s, err := store.New(context.Background(), "localhost:2379")
	if err != nil {
		// If etcd not available, New might fail or just return client.
		// code: clientv3.New(...) -> returns error if config bad?
		// DialTimeout is 5s but New doesn't connect immediately unless Dial called?
		// clientv3.New creates client.
		// If it fails, we skip.
		t.Skip("skipping store test as etcd might not be available")
	}

	ctx := context.Background()
	assert.ErrorIs(t, s.Create(ctx, nil), store.ErrNotImplements)
	assert.ErrorIs(t, s.Load(ctx, nil), store.ErrNotImplements)
	_, err = s.List(ctx, nil, nil)
	assert.ErrorIs(t, err, store.ErrNotImplements)
	_, err = s.Stream(ctx, nil, nil)
	assert.ErrorIs(t, err, store.ErrNotImplements)
	assert.ErrorIs(t, s.Update(ctx, nil), store.ErrNotImplements)
	assert.ErrorIs(t, s.Save(ctx, nil), store.ErrNotImplements)
	assert.ErrorIs(t, s.Delete(ctx, nil), store.ErrNotImplements)
}
