package settings_test

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bluemir/0xC0DE/internal/settings"
)

type TestConfig struct {
	Foo string
	Bar int
}

func TestSettings(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	manager := settings.New(db)
	ctx := context.Background()

	t.Run("Initialize New Config", func(t *testing.T) {
		conf := TestConfig{}
		err := manager.Load(ctx, "test-key", &conf, func() (any, error) {
			return TestConfig{Foo: "hello", Bar: 123}, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "hello", conf.Foo)
		assert.Equal(t, 123, conf.Bar)
	})

	t.Run("Load Existing Config", func(t *testing.T) {
		// New struct to ensure we load from DB
		conf := TestConfig{}
		// initFunc should not be called
		err := manager.Load(ctx, "test-key", &conf, func() (any, error) {
			t.Fatal("initFunc called when config exists")
			return nil, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "hello", conf.Foo)
		assert.Equal(t, 123, conf.Bar)
	})

	t.Run("Update Config", func(t *testing.T) {
		conf := TestConfig{Foo: "world", Bar: 456}
		err := manager.Save(ctx, "test-key", conf)
		require.NoError(t, err)

		// Verify load
		loaded := TestConfig{}
		err = manager.Load(ctx, "test-key", &loaded, func() (any, error) {
			t.Fatal("initFunc called")
			return nil, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "world", loaded.Foo)
		assert.Equal(t, 456, loaded.Bar)
	})
}
