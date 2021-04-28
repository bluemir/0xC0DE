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
			Kind:      "User",
			Id:        "hello",
			Namespace: "hello",
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
		Kind:      "User",
		Id:        "hello",
		Namespace: "hello",
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
			Kind:      "User",
			Id:        "t1",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})
	store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "t2",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})

	store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "t3",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "bb",
			},
		},
	})

	users, err := store.Find("User", "foo", map[string]string{
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
			Kind:      "User",
			Id:        "t1",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})
	store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "t2",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "aa",
				"state": "foo",
			},
		},
	})

	store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "t3",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "bb",
			},
		},
	})

	{
		users, err := store.Find("User", "foo", map[string]string{
			"owner": "aa",
			"state": "foo",
		})
		if err != nil {
			t.Error(err)
		}

		assert.Len(t, users, 1)
	}
	{
		users, err := store.Find("User", "foo", map[string]string{
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
			Kind:      "User",
			Id:        "hello",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "bot",
			},
		},
		Foo: "asdf",
	}); err != nil {
		t.Error(err)
	}

	if err := store.Load(&Metadata{
		Kind:      "User",
		Namespace: "foo",
	}, &User{}); err == nil {
		t.Errorf("must return error on empty id")
	}
}
func TestEmptyNamespace(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}
	if err := store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "hello",
			Namespace: "foo",
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
		Id:   "hello",
	}, &User{}); err == nil {
		t.Errorf("must return error on empty owner")
	}
}
func TestMultipleNamespace(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}

	users := []User{
		{
			Metadata: Metadata{
				Kind:      "User",
				Id:        "foo",
				Namespace: "t1",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			Metadata: Metadata{
				Kind:      "User",
				Id:        "foo",
				Namespace: "t2",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			Metadata: Metadata{
				Kind:      "User",
				Id:        "foo",
				Namespace: "t3",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	for _, u := range users {
		store.Save(&u)
	}

	{
		users, err := store.Find("User", "t1", map[string]string{
			"foo": "bar",
		})
		if err != nil {
			t.Error(err)
		}

		assert.Len(t, users, 1)
	}
}
func TestRev(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}

	store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "t1",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})
	if err := store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "t1",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "aa",
				"state": "foo",
			},
		},
	}); err == nil {
		t.Errorf("must return error on same Rev")
	}
}
func TestFindWithEmptyLabel(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Error(err)
	}

	store.Save(&User{
		Metadata: Metadata{
			Kind:      "User",
			Id:        "t1",
			Namespace: "foo",
			Labels: map[string]string{
				"owner": "aa",
			},
		},
	})
	{
		users, err := store.Find("User", "foo", nil)
		if err != nil {
			t.Error(err)
		}

		assert.Len(t, users, 1)
		assert.Equal(t, "foo", users[0].Namespace)
	}
}

type User struct {
	Metadata

	// Data...
	Foo string
}
