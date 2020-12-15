package server

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (server *Server) initDB() error {
	db, err := gorm.Open(sqlite.Open(server.conf.DBPath), &gorm.Config{})
	if err != nil {
		return err
	}
	server.db = db

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1)

	// TODO auto migrate
	return db.AutoMigrate()
}
