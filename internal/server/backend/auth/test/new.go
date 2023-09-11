package test

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth/store"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newManager() (*auth.Manager, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	if err != nil {
		return nil, err
	}

	s, err := store.Gorm(db, "")
	if err != nil {
		return nil, err
	}

	m, err := auth.New(s, "")
	if err != nil {
		return nil, err
	}
	return m, nil
}
