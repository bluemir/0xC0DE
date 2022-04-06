package auth

import (
	"time"
)

type User struct {
	Name      string `gorm:"primaryKey;size:256"`
	HashedKey string
	Salt      string
}
type Token struct {
	Username  string `gorm:"primaryKey;size:256"`
	HashedKey string `gorm:"primaryKey;size:256"`
	ExpiredAt *time.Time
}
