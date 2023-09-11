package gorm

import (
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

type DB = gorm.DB

func New(db *gorm.DB, salt string) (auth.AuthStore, error) {
	if err := db.AutoMigrate(
		&auth.User{},
		&auth.Token{},
		&auth.Group{},
		&auth.Role{},
		&auth.RoleBinding{},
	); err != nil {
		return nil, err
	}

	return &Store{
		db:   db,
		salt: salt,
	}, nil
}

type Store struct {
	db   *gorm.DB
	salt string
}
