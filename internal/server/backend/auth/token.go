package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/bluemir/0xC0DE/internal/util"
	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Token struct {
	Username  string    `gorm:"primaryKey;size:256"`
	Kind      TokenKind `gorm:"primaryKey;size:256;uniqueIndex:idx_tokens_external,priority:1"` // Secret 중 어느 것이 채워졌는지 가리키는 조회 키
	Index     int       `gorm:"primaryKey;size:256"`
	ExpiredAt *time.Time
	CreatedAt time.Time

	// ExternalID 는 바깥에서 온 식별자로 사용자를 역추적하는 조회 키다.
	// passkey 는 credential ID, OAuth 는 provider 가 준 사용자 ID 를 담는다.
	// Secret 안에 두면 인덱스를 걸 수 없어 Kind 와 마찬가지로 밖에 둔다.
	// 값의 유일함은 provider 안에서만 보장되므로 Kind 와 묶어 인덱스를 건다.
	// 사용자가 직접 증명하는 수단(password, access-key)에서는 비어 있다.
	ExternalID []byte `gorm:"uniqueIndex:idx_tokens_external,priority:2"`

	Secret Secret `gorm:"type:bytes;serializer:gob"`
}

// Secret 은 인증 수단별 자격증명이다. 해당하는 것 하나만 non-nil 이다.
// 종류마다 쓰는 필드가 달라 컬럼으로 펼치지 않고 한 컬럼에 직렬화한다.
type Secret struct {
	Password  *PasswordSecret
	AccessKey *AccessKeySecret
	Passkey   *PasskeySecret
	OAuth     *OAuthSecret // 어느 provider 인지는 Kind 가 말해준다
}

type PasswordSecret struct {
	HashedSecret []byte
}

type AccessKeySecret struct {
	HashedSecret []byte
}

type PasskeySecret struct {
	Credential []byte // webauthn.Credential JSON
	Label      string // 사용자가 붙인 이름

	// LastUsedAt 은 이 passkey 로 마지막에 로그인한 시각이다. 쓰인 적이 없으면 nil 이다.
	// 목록에서 어느 기기를 지워도 되는지 판단할 근거로 보여준다.
	// 로그인할 때마다 Credential 을 다시 쓰므로 여기 두면 쓰기가 늘지 않는다.
	// 대신 gob 안이라 DB 에서 직접 읽거나 정렬할 수는 없다.
	LastUsedAt *time.Time
}

// OAuthSecret 은 연결된 외부 계정의 프로필이다.
// 로그인에 필요한 값은 Kind 와 ExternalID 가 다 가지고 있고, 여기 있는 것은 보여주기 위한 것이다.
type OAuthSecret struct {
	Email       string
	Name        string
	RefreshedAt time.Time // 프로필을 마지막으로 provider 에서 가져온 시각
}

type TokenKind string

const (
	TokenKindPassword  TokenKind = "password"
	TokenKindAccessKey TokenKind = "access-key"
	TokenKindPasskey   TokenKind = "passkey"

	// OAuth provider 는 Kind 하나씩 차지한다. Secret.OAuth 는 공유한다.
	TokenKindGoogle TokenKind = "google"
	TokenKindGitHub TokenKind = "github"
)

// Validate 는 받은 비밀이 이 토큰의 것인지 본다.
//
// 실패는 이유가 무엇이든 ErrUnauthorized 를 섞어 올린다. 만료든 불일치든 부르는 쪽에는
// "이 비밀로는 통과 못 한다" 하나이고, errs 미들웨어가 그것을 보고 401 을 낸다.
// 섞지 않으면 bcrypt 의 에러가 그대로 올라가 500 이 된다.
//
// errors.Mark 를 쓰지 않는다. 마커는 cockroachdb 의 Is 를 거쳐야 보여서,
// 표준 errors.Is 를 쓰는 쪽(테스트의 assert.ErrorIs 를 포함해)에서는 묻힌다.
func (token *Token) Validate(unhashedSecret string) error {
	if token.ExpiredAt != nil && token.ExpiredAt.Before(time.Now()) {
		return errors.Join(ErrUnauthorized, errors.New("token is expired"))
	}
	hashedSecret := []byte(nil)
	switch {
	case token.Secret.Password != nil:
		hashedSecret = token.Secret.Password.HashedSecret
	case token.Secret.AccessKey != nil:
		hashedSecret = token.Secret.AccessKey.HashedSecret
	default:
		// passkey 는 서버와 나눠 가진 비밀이 없다. FinishPasskeyLogin 으로 검증한다.
		return ErrUnauthorized
	}

	// 통과한 경우를 감싸면 안 된다. errors.Join 은 두 번째가 nil 이어도
	// 첫 번째를 담은 non-nil 을 돌려줘서, 그대로 두면 모든 로그인이 실패한다.
	if err := bcrypt.CompareHashAndPassword(hashedSecret, []byte(unhashedSecret)); err != nil {
		return errors.Join(ErrUnauthorized, err)
	}
	return nil
}

// validateSecret 은 Kind 와 Secret 이 어긋나지 않는지 본다.
// Kind 는 조회용 인덱스일 뿐이라 둘이 따로 놀 수 있어 저장 직전에 맞춰 둔다.
func (token *Token) validateSecret() error {
	filled := 0
	if token.Secret.Password != nil {
		filled++
	}
	if token.Secret.AccessKey != nil {
		filled++
	}
	if token.Secret.Passkey != nil {
		filled++
	}
	if token.Secret.OAuth != nil {
		filled++
	}

	matched := false
	external := false // 바깥에서 온 식별자로 역추적되는 수단인가

	switch token.Kind {
	case TokenKindPassword:
		matched = token.Secret.Password != nil
	case TokenKindAccessKey:
		matched = token.Secret.AccessKey != nil
	case TokenKindPasskey:
		matched, external = token.Secret.Passkey != nil, true
	case TokenKindGoogle, TokenKindGitHub:
		matched, external = token.Secret.OAuth != nil, true
	default:
		return errors.Errorf("unknown token kind %q", token.Kind)
	}

	if filled != 1 || !matched {
		return errors.Errorf("secret does not match token kind %q", token.Kind)
	}

	if external != (len(token.ExternalID) > 0) {
		return errors.Errorf("external id does not match token kind %q", token.Kind)
	}

	return nil
}

func (m *Manager) IssueToken(ctx context.Context, username string, kind TokenKind, unhashedSecret string, opts ...TokenOpt) (*Token, error) {
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(unhashedSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	token := &Token{
		Username: username,
		Kind:     kind,
	}
	switch kind {
	case TokenKindPassword:
		token.Secret.Password = &PasswordSecret{HashedSecret: hashedSecret}
	case TokenKindAccessKey:
		token.Secret.AccessKey = &AccessKeySecret{HashedSecret: hashedSecret}
	default:
		return nil, errors.Errorf("token kind %q is not issued with a secret", kind)
	}

	for _, fn := range opts {
		fn(token)
	}

	if err := m.createToken(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// createToken 은 같은 (username, kind) 안에서 다음 index 를 붙여 token 을 저장한다.
func (m *Manager) createToken(ctx context.Context, token *Token) error {
	if err := token.validateSecret(); err != nil {
		return err
	}

	tx := m.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	lastToken := Token{}

	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("username = ? AND kind = ?", token.Username, token.Kind).
		Order("`index` desc").
		First(&lastToken)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.WithStack(result.Error)
	}
	if result.RowsAffected > 0 {
		token.Index = lastToken.Index + 1
	} else {
		token.Index = 0
	}

	if err := tx.Create(token).Error; err != nil {
		return errors.WithStack(err)
	}
	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	return nil
}
func (m *Manager) UpdatePassword(ctx context.Context, username string, unhashedPassword string) error {
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(unhashedPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	// Assignments 에 값을 직접 넣으면 serializer 를 거치지 않는다.
	// AssignmentColumns 는 INSERT 에 쓰인 값을 그대로 가져다 쓴다.
	if err := m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}, {Name: "kind"}, {Name: "index"}},
		DoUpdates: clause.AssignmentColumns([]string{"secret"}),
	}).Create(&Token{
		Username: username,
		Kind:     TokenKindPassword,
		Index:    0,
		Secret:   Secret{Password: &PasswordSecret{HashedSecret: hashedSecret}},
	}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func (m *Manager) GenerateAccessKey(ctx context.Context, username string, opts ...TokenOpt) (*Token, string, error) {
	unhashedSecret := util.RandomString(32)

	t, err := m.IssueToken(ctx, username, TokenKindAccessKey, unhashedSecret, opts...)
	if err != nil {
		return nil, "", err
	}

	// {username}.{index}.{secret}

	return t, fmt.Sprintf("%s.%d.%s", username, t.Index, unhashedSecret), nil
}
func (m *Manager) GetToken(ctx context.Context, username string, kind TokenKind, index int) (*Token, error) {
	token := Token{}
	if err := m.db.WithContext(ctx).Where(Token{
		Username: username,
		Kind:     kind,
		Index:    index,
	}).Take(&token).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &token, nil
}
func (m *Manager) ListToken(ctx context.Context, username string) ([]Token, error) {
	tokens := []Token{}

	if err := m.db.WithContext(ctx).Where(Token{Username: username}).Find(&tokens).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return tokens, nil
}
func (m *Manager) RevokeToken(ctx context.Context, username string, kind TokenKind, index int) error {
	if err := m.db.WithContext(ctx).Where(Token{
		Username: username,
		Kind:     kind,
		Index:    index,
	}).Delete(Token{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type TokenOpt func(*Token)

func ExpiredAt(t time.Time) func(*Token) {
	return func(token *Token) {
		token.ExpiredAt = &t
	}
}
func ExpiredAfter(d time.Duration) func(*Token) {
	return func(token *Token) {
		t := time.Now().Add(d)
		token.ExpiredAt = &t
	}
}
