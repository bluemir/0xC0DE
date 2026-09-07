package auth

import (
	"fmt"
	"time"

	"github.com/bluemir/0xC0DE/internal/util"
	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Token struct {
	Username     string    `gorm:"primaryKey;size:256"`
	Kind         TokenKind `gorm:"primaryKey;size:256"` // password, access_keys, google, github...
	Index        int       `gorm:"primaryKey;size:256"`
	HashedSecret []byte
	ExpiredAt    *time.Time
	CreatedAt    time.Time

	// passkey 전용. Kind == TokenKindPasskey 일 때만 채워진다.
	// 공개키는 비밀이 아니므로 HashedSecret 대신 별도 컬럼에 담는다.
	CredentialID []byte `gorm:"uniqueIndex"` // 로그인 시 사용자를 찾는 키
	Credential   []byte // webauthn.Credential JSON
	Label        string // 사용자가 붙인 이름
}

type TokenKind string

const (
	TokenKindPassword  TokenKind = "password"
	TokenKindAccessKey TokenKind = "access-key"
	TokenKindPasskey   TokenKind = "passkey"
)

func (token *Token) Validate(unhashedSecret string) error {
	if token.ExpiredAt != nil && token.ExpiredAt.Before(time.Now()) {
		return errors.New("token is expired") // TODO
	}
	return bcrypt.CompareHashAndPassword(token.HashedSecret, []byte(unhashedSecret))
}

func (m *Manager) IssueToken(username string, kind TokenKind, unhashedSecret string, opts ...TokenOpt) (*Token, error) {
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(unhashedSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	token := &Token{
		Username:     username,
		Kind:         kind,
		HashedSecret: hashedSecret,
	}
	for _, fn := range opts {
		fn(token)
	}

	if err := m.createToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// createToken 은 같은 (username, kind) 안에서 다음 index 를 붙여 token 을 저장한다.
func (m *Manager) createToken(token *Token) error {
	tx := m.db.Begin()
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
func (m *Manager) UpdatePassword(username string, unhashedPassword string) error {
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(unhashedPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := m.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}, {Name: "kind"}, {Name: "index"}},
		DoUpdates: clause.Assignments(map[string]any{"hashed_secret": hashedSecret}),
	}).Create(&Token{
		Username:     username,
		Kind:         TokenKindPassword,
		Index:        0,
		HashedSecret: hashedSecret,
	}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
func (m *Manager) GenerateAccessKey(username string, opts ...TokenOpt) (*Token, string, error) {
	unhashedSecret := util.RandomString(32)

	t, err := m.IssueToken(username, TokenKindAccessKey, unhashedSecret, opts...)
	if err != nil {
		return nil, "", err
	}

	// {username}.{index}.{secret}

	return t, fmt.Sprintf("%s.%d.%s", username, t.Index, unhashedSecret), nil
}
func (m *Manager) GetToken(username string, kind TokenKind, index int) (*Token, error) {
	token := Token{}
	if err := m.db.Where(Token{
		Username: username,
		Kind:     kind,
		Index:    index,
	}).Take(&token).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &token, nil
}
func (m *Manager) ListToken(username string) ([]Token, error) {
	tokens := []Token{}

	if err := m.db.Where(Token{Username: username}).Find(&tokens).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return tokens, nil
}
func (m *Manager) RevokeToken(username string, kind TokenKind, index int) error {
	if err := m.db.Where(Token{
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
