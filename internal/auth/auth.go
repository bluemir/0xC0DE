package auth

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"
)

func New(db *gorm.DB, salt string) (*Manager, error) {
	if err := db.AutoMigrate(
		&User{},
		&Token{},
	); err != nil {
		return nil, err
	}

	return &Manager{db, salt}, nil
}

type Manager struct {
	db   *gorm.DB
	salt string
}
type User struct {
	Name      string `gorm:"primaryKey;size:256"`
	Salt      string
	HashedKey string
}
type Token struct {
	Key      string `gorm:"primaryKey;size:256"`
	UserName string
	ExpireAt time.Time
}

func (m *Manager) Default(name, unhashedKey string) (*Token, error) {
	user := &User{}
	if err := m.db.Where(&User{Name: name}).Take(user).Error; err != nil {
		return nil, err
	}

	hashedKey := hash(unhashedKey, user.Salt, m.salt)
	if user.HashedKey != hashedKey {
		return nil, ErrKeyNotMatched
	}

	token := &Token{
		UserName: user.Name,
		Key:      xid.New().String(),
	}
	// XXX expire previous one?
	if err := m.db.Save(token).Error; err != nil {
		return nil, err
	}

	// make new token
	return token, nil
}
func (m *Manager) GetToken(tokenKey string) (*Token, error) {
	token := &Token{}
	if err := m.db.Where(&Token{Key: tokenKey}).Take(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}
func (m *Manager) Register(username, unhashedKey string) (*User, error) {
	salt := xid.New().String()
	user := &User{
		Name:      username,
		HashedKey: hash(unhashedKey, salt, m.salt),
		Salt:      salt,
	}
	if err := m.db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
