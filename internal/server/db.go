package server

import (
	"context"

	"github.com/cockroachdb/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initializeDB(ctx context.Context, dbpath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	return db, nil
}
