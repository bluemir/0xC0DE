package auth

import (
	"github.com/bluemir/functional/v2"
	"github.com/cockroachdb/errors"
	"github.com/rs/xid"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

type User struct {
	Name string `gorm:"primaryKey;size:256" json:"name" expr:"name"`
	Salt string `json:"-"`
	// Handle 은 WebAuthn user handle 이다.
	// authenticator 에 노출되는 값이라 사용자 이름과 분리한다.
	Handle []byte  `gorm:"uniqueIndex" json:"-"`
	Groups []Group `gorm:"many2many:members;" json:"groups"`
	Labels Labels  `gorm:"type:bytes;serializer:gob" json:"labels" expr:"labels"`
}

func (m *Manager) Register(username, unhashedPassword string, opts ...CreateUserOption) (*User, *Token, error) {
	user, err := m.CreateUser(username, opts...)
	if err != nil {
		return nil, nil, err
	}

	accessKey, err := m.IssueToken(username, TokenKindPassword, unhashedPassword)
	if err != nil {
		return user, nil, err
	}

	return user, accessKey, nil
}

type CreateUserOption func(u *User)

// withHandle 은 이미 만들어 둔 WebAuthn user handle 을 그대로 쓴다.
// passkey 가입은 ceremony 를 시작할 때 handle 을 먼저 만들기 때문에 필요하다.
func withHandle(handle []byte) func(*User) {
	return func(u *User) {
		u.Handle = handle
	}
}

func WithGroup(groups ...string) func(*User) {
	return func(u *User) {
		u.Groups = functional.SliceMap(groups, func(g string) Group {
			return Group{
				Name: g,
			}
		})
	}
}

func (m *Manager) CreateUser(username string, opts ...CreateUserOption) (*User, error) {
	handle, err := newPasskeyHandle()
	if err != nil {
		return nil, err
	}

	u := User{
		Name:   username,
		Salt:   xid.New().String(),
		Handle: handle,
		Groups: []Group{},
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

	if err := m.db.Preload("Groups").Where(User{Name: username}).Take(&u).Error; err != nil {
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
	if option.Limit == 0 {
		option.Limit = 20
	}
	users := []User{}

	if err := m.db.Preload("Groups").Offset(option.Offset).Limit(option.Limit).Find(&users).Error; err != nil {
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
	for _, g := range u.Groups {
		ret = append(ret, Subject{
			Kind: KindGroup,
			Name: g.Name,
		})
	}
	return ret
}
