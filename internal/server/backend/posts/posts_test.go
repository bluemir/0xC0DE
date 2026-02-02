package posts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/bluemir/0xC0DE/internal/server/backend/posts"
)

func newTestManager(t *testing.T) *posts.Manager {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	ctx := context.Background()
	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	conf := &posts.Config{} // Empty config
	m, err := posts.New(ctx, conf, db, hub)
	require.NoError(t, err)

	return m
}

func TestCreateAndList(t *testing.T) {
	m := newTestManager(t)

	// Create
	ctx := context.Background()
	post, err := m.Create(ctx, "Content 1")
	assert.NoError(t, err)
	assert.NotEmpty(t, post.Id)
	assert.Equal(t, "Content 1", post.Message)

	// List
	list, err := m.List(ctx, meta.Limit(10))
	assert.NoError(t, err)
	assert.Len(t, list.Items, 1)
	assert.Equal(t, "Content 1", list.Items[0].Message)

	// query find
	query := posts.Query{}
	foundList, err := m.FindWithOption(ctx, query, meta.ListOption{Limit: 1})
	assert.NoError(t, err)
	assert.Positive(t, foundList.Total)
	assert.NotEmpty(t, foundList.Items)
}
