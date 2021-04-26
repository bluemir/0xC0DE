package store

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initStore() (*Store, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if rdb, err := db.DB(); err != nil {
		return nil, err
	} else {
		rdb.SetMaxOpenConns(1)
	}
	store, err := New(db)
	if err != nil {
		return nil, err
	}
	return store, nil
}
func TestMain(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}

	if err := store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "hello",
			Labels: map[string]string{
				"owner": "bot",
			},
		},
		Foo: "asdf",
	}); err != nil {
		t.Error(err)
	}

	tUser := &User{}
	err = store.Load(&Metadata{
		Kind: "User",
		Id:   "hello",
	}, tUser)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "asdf", tUser.Foo)
	assert.Equal(t, tUser.GetMetadata().Rev, 1)

	if err := store.Save(tUser); err != nil {
		t.Error(err)
	}

	tUser.Rev = 1
	tUser.Foo = "xxx"

	if err := store.Save(tUser); err == nil {
		t.Error("must faild")
	}
}
func TestLabel(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}

	store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "t1",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})
	store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "t2",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})

	store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "t3",
			Labels: map[string]string{
				"owner": "bb",
			},
		},
	})

	users, err := store.Find(map[string]string{
		"owner": "aa",
	})
	if err != nil {
		t.Error(err)
	}

	assert.Len(t, users, 2)
	for _, user := range users {
		assert.Equal(t, user.Labels["owner"], "aa")
	}
}
func TestMultipleLabels(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}

	store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "t1",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})
	store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "t2",
			Labels: map[string]string{
				"owner": "aa",
				"state": "foo",
			},
		},
	})

	store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "t3",
			Labels: map[string]string{
				"owner": "bb",
			},
		},
	})

	{
		users, err := store.Find(map[string]string{
			"owner": "aa",
			"state": "foo",
		})
		if err != nil {
			t.Error(err)
		}

		assert.Len(t, users, 1)
	}
	{
		users, err := store.Find(map[string]string{
			"owner": "aa",
		})
		if err != nil {
			t.Error(err)
		}

		assert.Len(t, users, 2)
	}
}
func TestEmptyID(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}
	if err := store.Save(&User{
		Metadata: Metadata{
			Kind: "User",
			Id:   "hello",
			Labels: map[string]string{
				"owner": "bot",
			},
		},
		Foo: "asdf",
	}); err != nil {
		t.Error(err)
	}

	if err := store.Load(&Metadata{
		Kind: "User",
	}, &User{}); err == nil {
		t.Errorf("must return error on empty id")
	}
}

type User struct {
	Metadata

	// Data...
	Foo string
}
