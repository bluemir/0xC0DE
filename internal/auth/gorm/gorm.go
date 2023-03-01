package gorm

import (
	"gorm.io/gorm"
)

type Manager struct {
	db    *gorm.DB
	roles map[string]Role
	salt  string
}
