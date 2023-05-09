package gorm

import (
	"time"

	"github.com/bluemir/0xC0DE/internal/auth"
	"gorm.io/gorm"
)

func (m *Manager) IssueToken(username, unhashedKey string, expiredAt *time.Time) error {
	user := &User{}
	if err := m.db.Where(&User{Name: username}).Take(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return auth.ErrUnauthorized
		}
		return err
	}

	token := &Token{
		Username:  username,
		HashedKey: auth.Hash(unhashedKey, user.Salt, m.salt),
		ExpiredAt: expiredAt,
	}

	if err := m.db.Save(token).Error; err != nil {
		return err
	}

	return nil
}

func (m *Manager) RevokeToken(username, unhashedKey string) error {
	user := &User{}
	if err := m.db.Where(&User{Name: username}).Take(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return auth.ErrUnauthorized
		}
		return err
	}

	token := &Token{
		Username:  username,
		HashedKey: auth.Hash(unhashedKey, user.Salt, m.salt),
	}
	/*
		XXX has bug. see https://github.com/go-gorm/gorm/issues/4879
		if err := m.db.Delete(token).Error; err != nil {
			return err
		}
	*/
	if err := m.db.Model(token).Where(token).Delete(struct {
		UserName  string
		HashedKey string
	}{}).Error; err != nil {
		return err
	}

	return nil
}
func (m *Manager) RevokeExpiredToken() error {
	if err := m.db.Model(&Token{}).Where("expired_at < ?", time.Now()).Delete(&struct{}{}).Error; err != nil {
		return err
	}
	return nil
}
