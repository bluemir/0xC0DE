package auth

import (
	"context"
	"time"

	"github.com/bluemir/functional/v2"
	"github.com/cockroachdb/errors"
	"github.com/rs/xid"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

type User struct {
	Name string `gorm:"primaryKey;size:256" json:"name" expr:"name"`
	// Email 은 계정 복구 메일을 받을 주소다. 비어 있을 수 있다.
	// 같은 주소를 여러 계정이 쓸 수 있어 uniqueIndex 가 아니다.
	Email string `gorm:"index;size:256" json:"email" expr:"email"`
	// EmailVerifiedAt 은 확인 링크를 눌러 이 주소를 증명한 시각이다. 아직이면 nil 이다.
	// 복구를 막지는 않는다. 확인하지 않은 주소라는 것을 화면에 알려 주는 데 쓴다.
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	Salt            string     `json:"-"`
	// Handle 은 WebAuthn user handle 이다.
	// authenticator 에 노출되는 값이라 사용자 이름과 분리한다.
	Handle []byte  `gorm:"uniqueIndex" json:"-"`
	Groups []Group `gorm:"many2many:members;" json:"groups"`
	Labels Labels  `gorm:"type:bytes;serializer:gob" json:"labels" expr:"labels"`
}

func (m *Manager) Register(ctx context.Context, username, unhashedPassword string, opts ...CreateUserOption) (*User, *Token, error) {
	user, err := m.CreateUser(ctx, username, opts...)
	if err != nil {
		return nil, nil, err
	}

	accessKey, err := m.IssueToken(ctx, username, TokenKindPassword, unhashedPassword)
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

func WithEmail(email string) func(*User) {
	return func(u *User) {
		u.Email = email
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

func (m *Manager) CreateUser(ctx context.Context, username string, opts ...CreateUserOption) (*User, error) {
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
	if err := m.db.WithContext(ctx).Create(&u).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &u, nil
}
func (m *Manager) GetUser(ctx context.Context, username string) (*User, error) {
	u := User{}

	if err := m.db.WithContext(ctx).Preload("Groups").Where(User{Name: username}).Take(&u).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &u, nil
}

// FindUserByEmail 은 이메일로 사용자를 찾는다.
// 같은 주소를 쓰는 계정이 여럿이면 가장 먼저 만들어진 것을 돌려준다.
func (m *Manager) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}

	if email == "" {
		return nil, errors.WithStack(gorm.ErrRecordNotFound)
	}

	if err := m.db.WithContext(ctx).Preload("Groups").Where(User{Email: email}).Take(&u).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return &u, nil
}

// UpdateEmail 은 이메일만 바꾼다.
// UpdateUser(Save) 는 구조체 전체를 덮어써서 Groups 나 Labels 까지 건드린다.
//
// 주소가 달라지면 확인 상태를 지우고 이전 주소로 보낸 확인 링크도 폐기한다.
// 같은 주소를 다시 저장하는 것은 아무것도 바꾸지 않는다. 확인된 주소를 저장 버튼 한 번에
// 미확인으로 되돌릴 이유가 없다.
func (m *Manager) UpdateEmail(ctx context.Context, username string, email string) error {
	user, err := m.GetUser(ctx, username)
	if err != nil {
		return err
	}
	if user.Email == email {
		return nil
	}

	// map 으로 넘겨야 zero value 인 nil 도 같이 쓰인다.
	if err := m.db.WithContext(ctx).Model(&User{}).Where(User{Name: username}).Updates(map[string]any{
		"email":             email,
		"email_verified_at": nil,
	}).Error; err != nil {
		return errors.WithStack(err)
	}

	return m.RevokeVerification(ctx, username)
}

func (m *Manager) ListUser(ctx context.Context, opts ...meta.ListOptionFn) ([]User, error) {
	opt := meta.ListOption{}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.ListUserWithOption(ctx, opt)
}
func (m *Manager) ListUserWithOption(ctx context.Context, option meta.ListOption) ([]User, error) {
	if option.Limit == 0 {
		option.Limit = 20
	}
	users := []User{}

	if err := m.db.WithContext(ctx).Preload("Groups").Offset(option.Offset).Limit(option.Limit).Find(&users).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return users, nil
}
func (m *Manager) UpdateUser(ctx context.Context, user *User) error {
	if err := m.db.WithContext(ctx).Save(user).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// EventUserDeleted 는 계정이 지워진 뒤에 발행된다.
// 이 시점에 users 행도, auth 가 가진 딸린 행도 모두 없다.
//
// 지워진 User 를 통째로 담는다. 이름만 담으면 받는 쪽이 더 알아볼 방법이 없다.
// auth 밖에서 사용자에 딸린 데이터를 가진 쪽이 이것을 보고 자기 것을 치운다.
type EventUserDeleted struct {
	User User
}

// DeleteUser 는 계정과 auth 가 가진 딸린 행을 한 트랜잭션에서 지운다.
//
// 딸린 행(토큰, 복구·확인 비밀, 역할 배정, 그룹 소속)을 users 행과 같이 지운다.
// 특히 tokens 가 남으면 같은 이름으로 다시 가입했을 때 죽은 계정의 비밀번호와
// passkey 가 새 계정에 그대로 붙는다. 하나라도 실패하면 계정도 지워지지 않는다.
//
// 이벤트는 커밋된 뒤에 보낸다. 되돌아간 삭제를 알릴 수는 없고, 트랜잭션이 열린 채
// 보내면 받는 쪽이 DB 를 치다 멈춘다. 커넥션 하나로 도는 구성에서는 그대로 걸린다.
//
// 알리는 시점에 치울 것은 이미 다 치웠으므로 결과를 보지 않는 Notify 다.
// 받는 쪽은 auth 밖에서 사용자에 딸린 데이터를 가진 백엔드이지, 이 삭제의 일부가 아니다.
func (m *Manager) DeleteUser(ctx context.Context, username string) error {
	user, err := m.GetUser(ctx, username)
	if err != nil {
		return err
	}

	tx := m.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Delete(User{Name: username}).Error; err != nil {
		return errors.WithStack(err)
	}
	if err := deleteUserRows(tx, username); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	m.events.Notify(ctx, EventUserDeleted{User: *user})

	return nil
}

// deleteUserRows 는 username 에 딸린 행을 지운다.
// users 행을 지우는 트랜잭션 안에서 불린다.
func deleteUserRows(tx *gorm.DB, username string) error {
	// 로그인 수단 전부(password, access-key, passkey, oauth)
	if err := tx.Where(Token{Username: username}).Delete(Token{}).Error; err != nil {
		return errors.WithStack(err)
	}

	// 계정 복구와 이메일 확인에 쓰던 일회용 비밀
	if err := tx.Where(Recovery{Username: username}).Delete(Recovery{}).Error; err != nil {
		return errors.WithStack(err)
	}
	if err := tx.Where(Verification{Username: username}).Delete(Verification{}).Error; err != nil {
		return errors.WithStack(err)
	}

	// 역할 배정. Subject 가 embedded 라 구조체 조건이 그대로 안 먹어 컬럼으로 적는다
	if err := tx.Where("subject_kind = ? AND subject_name = ?", KindUser, username).
		Delete(Assign{}).Error; err != nil {
		return errors.WithStack(err)
	}

	// 그룹 소속. many2many 조인 테이블이라 대응하는 모델이 없다
	if err := tx.Exec("DELETE FROM members WHERE user_name = ?", username).Error; err != nil {
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
