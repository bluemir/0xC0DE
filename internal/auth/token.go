package auth

import (
	"time"

	"gorm.io/gorm"
)

func (m *Manager) IssueToken(username, unhashedKey string, expiredAt *time.Time) error {
	user := &User{}
	if err := m.db.Where(&User{Name: username}).Take(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUnauthroized
		}
		return err
	}

	token := &Token{
		Username:  username,
		HashedKey: hash(unhashedKey, user.Salt, m.salt),
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
			return ErrUnauthroized
		}
		return err
	}

	token := &Token{
		Username:  username,
		HashedKey: hash(unhashedKey, user.Salt, m.salt),
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
