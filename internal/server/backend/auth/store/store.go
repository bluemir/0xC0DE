package store

import (
	gorm "gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	gormStore "github.com/bluemir/0xC0DE/internal/server/backend/auth/store/gorm"
)

//facade

func Gorm(db *gorm.DB, salt string) (auth.AuthStore, error) {
	return gormStore.New(db, salt)
}
