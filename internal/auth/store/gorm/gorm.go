package gorm

import (
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/auth"
)

type DB = gorm.DB

func New(db *gorm.DB, salt string) (auth.AuthStore, error) {
	if err := db.AutoMigrate(
		&auth.User{},
		&auth.Token{},
		&auth.Group{},
		&Role{},
		&RoleBinding{},
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
