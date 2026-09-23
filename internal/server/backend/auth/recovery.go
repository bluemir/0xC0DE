package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// RecoveryLifetime 은 복구 링크가 살아 있는 시간이다.
	RecoveryLifetime = 15 * time.Minute
	// RecoveryResendInterval 은 복구 메일을 다시 보낼 수 있을 때까지 기다리는 시간이다.
	RecoveryResendInterval = 15 * time.Minute

	recoverySecretSize = 32
)

// ErrRecoveryTooSoon 은 직전 복구 메일을 보낸 지 얼마 안 됐을 때다.
// 호출측은 이것을 사용자에게 그대로 드러내지 않는다. 계정 존재 여부가 새기 때문이다.
var ErrRecoveryTooSoon = errors.Errorf("recovery mail was sent too recently")

// Recovery 는 계정 복구 링크에 담긴 일회용 비밀이다.
//
// Token 에 섞지 않는다. 이것은 로그인 수단이 아니라 인증 수단을 다시 세우기 위한 임시
// 통행증이고, Token 에 넣으면 ListToken 에 딸려 나오는 데다 Kind 마다 index 가 늘어나
// 요청할 때마다 행이 쌓인다. 사용자당 한 행만 두고 재발급하면 덮어쓴다.
type Recovery struct {
	Username     string `gorm:"primaryKey;size:256"`
	HashedSecret []byte
	ExpiredAt    time.Time
	CreatedAt    time.Time
}

// IssueRecovery 는 복구 비밀을 새로 발급한다.
// 돌려주는 값은 해시 전의 원본이고, 이 순간이 지나면 다시 알 수 없다.
//
// 직전 발급으로부터 RecoveryResendInterval 이 지나지 않았으면 ErrRecoveryTooSoon 이다.
func (m *Manager) IssueRecovery(username string) (string, error) {
	secret, err := newRecoverySecret()
	if err != nil {
		return "", err
	}

	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.WithStack(err)
	}

	now := time.Now()

	tx := m.db.Begin()
	defer tx.Rollback()

	last := Recovery{}
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(Recovery{Username: username}).
		Take(&last)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", errors.WithStack(result.Error)
	}
	if result.RowsAffected > 0 && last.CreatedAt.Add(RecoveryResendInterval).After(now) {
		return "", ErrRecoveryTooSoon
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{"hashed_secret", "expired_at", "created_at"}),
	}).Create(&Recovery{
		Username:     username,
		HashedSecret: hashedSecret,
		ExpiredAt:    now.Add(RecoveryLifetime),
		CreatedAt:    now,
	}).Error; err != nil {
		return "", errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return "", errors.WithStack(err)
	}

	return secret, nil
}

// ValidateRecovery 는 복구 링크의 비밀을 검증하고 그 계정을 돌려준다.
// 없는 계정, 만료, 비밀 불일치를 모두 ErrUnauthorized 하나로 답한다.
func (m *Manager) ValidateRecovery(username string, unhashedSecret string) (*User, error) {
	recovery := Recovery{}
	if err := m.db.Where(Recovery{Username: username}).Take(&recovery).Error; err != nil {
		return nil, ErrUnauthorized
	}

	if recovery.ExpiredAt.Before(time.Now()) {
		return nil, ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword(recovery.HashedSecret, []byte(unhashedSecret)); err != nil {
		return nil, ErrUnauthorized
	}

	user, err := m.GetUser(username)
	if err != nil {
		return nil, ErrUnauthorized
	}
	return user, nil
}

// RevokeRecovery 는 복구 비밀을 폐기한다. 없어도 에러가 아니다.
func (m *Manager) RevokeRecovery(username string) error {
	if err := m.db.Where(Recovery{Username: username}).Delete(Recovery{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// newRecoverySecret 은 링크에 담을 비밀을 만든다.
// util.RandomString 은 math/rand 라 예측할 수 있어서 쓰지 않는다.
func newRecoverySecret() (string, error) {
	buf := make([]byte, recoverySecretSize)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.WithStack(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
