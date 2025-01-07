package auth

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

type User struct {
	Name   string `gorm:"primaryKey;size:256" json:"name" expr:"name"`
	Salt   string `json:"-"`
	Groups Set    `gorm:"type:bytes;serializer:gob" json:"groups"`
	Labels Labels `gorm:"type:bytes;serializer:gob" json:"labels" expr:"labels"`
}

func (m *Manager) Register(username, unhashedKey string, opts ...CreateUserOption) (*User, *Token, error) {
	user, err := m.CreateUser(username, opts...)
	if err != nil {
		return nil, nil, err
	}

	token, err := m.IssueToken(username, unhashedKey)
	if err != nil {
		return user, nil, err
	}

	return user, token, nil
}

type CreateUserOption func(u *User)

func WithGroup(groups ...string) func(*User) {
	return func(u *User) {
		u.Groups = setFromArray(groups)
	}
}

func (m *Manager) CreateUser(username string, opts ...CreateUserOption) (*User, error) {
	u := User{
		Name:   username,
		Salt:   xid.New().String(),
		Groups: Set{},
	}
	for _, fn := range opts {
		fn(&u)
	}
	if err := m.db.Create(&u).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &u, nil
}
func (m *Manager) GetUser(username string) (*User, error) {
	u := User{}

	if err := m.db.Where(User{Name: username}).Take(&u).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &u, nil
}
func (m *Manager) ListUser(opts ...meta.ListOptionFn) ([]User, error) {
	opt := meta.ListOption{}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.ListUserWithOption(opt)
}
func (m *Manager) ListUserWithOption(option meta.ListOption) ([]User, error) {
	users := []User{}

	if err := m.db.Offset(option.Offset).Limit(option.Limit).Find(users).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return users, nil
}
func (m *Manager) UpdateUser(user *User) error {
	if err := m.db.Save(user).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func (m *Manager) DeleteUser(username string) error {
	if err := m.db.Delete(User{Name: username}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func (u *User) Subjects() []Subject {

	if u == nil {
		return []Subject{
			{
				Kind: KindGuest,
			},
		}
	}
	ret := []Subject{
		{
			Kind: KindUser,
			Name: u.Name,
		},
	}
	for g := range u.Groups {
		ret = append(ret, Subject{
			Kind: KindGroup,
			Name: g,
		})
	}
	return ret
}
