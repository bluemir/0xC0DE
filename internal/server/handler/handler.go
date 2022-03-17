package handler

import "gorm.io/gorm"

func New(db *gorm.DB) (*Handler, error) {
	return &Handler{
		db: db,
	}, nil
}

type Handler struct {
	db *gorm.DB
}
