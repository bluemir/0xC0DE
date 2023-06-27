package server

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	// TODO auto migrate
	if err := db.AutoMigrate(); err != nil {
		return nil, err
	}
	return db, err
}
