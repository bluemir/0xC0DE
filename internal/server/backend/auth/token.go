package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pkg/errors"
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
}

type TokenKind string

const (
	TokenKindPassword  TokenKind = "password"
	TokenKindAccessKey TokenKind = "access-key"
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

	tx := m.db.Begin()
	defer tx.Rollback()

	lastToken := Token{}

	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("username = ? AND kind = ?", username, kind).
		Order("index desc").
		First(&lastToken)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.WithStack(result.Error)
	}
	if result.RowsAffected > 0 {
		token.Index = lastToken.Index + 1
	} else {
		token.Index = 0
	}

	if err := tx.Create(token).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return token, nil
}
func (m *Manager) UpdatePassword(username string, unhashedPassword string) error {
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(unhashedPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := m.db.Save(&Token{
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
	unhashedSecret, err := generateRandomString(32)

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
func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
