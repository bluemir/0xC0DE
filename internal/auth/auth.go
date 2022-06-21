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
		&RoleBinding{},
	); err != nil {
		return nil, err
	}

	return &Manager{
		db:    db,
		roles: map[string]Role{},
		salt:  salt,
	}, nil
}

type Manager struct {
	db    *gorm.DB
	roles map[string]Role
	salt  string
}

func (m *Manager) Default(username, unhashedKey string) (*User, error) {
	user := &User{}
	if err := m.db.Where(&User{
		Name: username,
	}).Take(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUnauthroized
		}
		return nil, err
	}

	token := &Token{}
	if err := m.db.Where(&Token{
		Username:  username,
		HashedKey: hash(unhashedKey, user.Salt, m.salt),
	}).Take(token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUnauthroized
		}
		return nil, err
	}

	if token.ExpiredAt == nil {
		return user, nil
	}

	if token.ExpiredAt.Before(time.Now()) {
		m.RevokeToken(username, unhashedKey)
		return nil, ErrUnauthroized
	}

	return user, nil
}

func (m *Manager) Register(username, unhashedKey string) (*User, error) {
	salt := xid.New().String()

	user := &User{
		Name: username,
		Salt: salt,
	}
	if err := m.db.Save(user).Error; err != nil {
		return nil, err
	}

	if err := m.db.Save(&Token{
		Username:  username,
		HashedKey: hash(unhashedKey, salt, m.salt),
	}).Error; err != nil {
		return nil, err
	}
	return user, nil
}
