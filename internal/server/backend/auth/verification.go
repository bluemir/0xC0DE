package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// VerificationLifetime 은 이메일 확인 링크가 살아 있는 시간이다.
	// 복구 링크(RecoveryLifetime)보다 길다. 복구는 로그인 못 하는 사람이 지금 당장 누르는
	// 것이지만, 확인 메일은 나중에 메일함을 열었을 때 눌러도 되기 때문이다.
	VerificationLifetime = 24 * time.Hour
	// VerificationResendInterval 은 확인 메일을 다시 보낼 수 있을 때까지 기다리는 시간이다.
	// 복구(RecoveryResendInterval)보다 짧다. 이쪽은 로그인한 사용자만 보낼 수 있고,
	// "메일이 안 왔으니 다시" 가 흔한 흐름이다.
	VerificationResendInterval = 5 * time.Minute

	verificationSecretSize = 32
)

// ErrVerificationTooSoon 은 직전 확인 메일을 보낸 지 얼마 안 됐을 때다.
var ErrVerificationTooSoon = errors.Errorf("verification mail was sent too recently")

// Verification 은 이메일 확인 링크에 담긴 일회용 비밀이다.
//
// Recovery 와 나란히 두고 합치지 않는다. 둘은 수명도 재발송 간격도 다르고,
// 복구 중에 이메일을 바꾸면 한 모델에서 서로를 덮어쓴다.
//
// Email 을 같이 담는다. 확인 메일을 보낸 뒤 주소를 또 바꾸면 날아간 링크는
// 이제 다른 주소를 가리키므로, 확인할 때 지금 주소와 같은지 본다.
type Verification struct {
	Username     string `gorm:"primaryKey;size:256"`
	Email        string `gorm:"size:256"`
	HashedSecret []byte
	ExpiredAt    time.Time
	CreatedAt    time.Time
}

// IssueVerification 은 이메일 확인 비밀을 새로 발급한다.
// 돌려주는 값은 해시 전의 원본이고, 이 순간이 지나면 다시 알 수 없다.
//
// 직전 발급으로부터 VerificationResendInterval 이 지나지 않았으면 ErrVerificationTooSoon 이다.
// 단 주소가 달라졌으면 간격과 상관없이 새로 발급한다. 오타를 고친 사용자를 기다리게 할 이유가 없다.
func (m *Manager) IssueVerification(username string, email string) (string, error) {
	if email == "" {
		return "", errors.Errorf("cannot verify empty email")
	}

	secret, err := newVerificationSecret()
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

	last := Verification{}
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(Verification{Username: username}).
		Take(&last)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", errors.WithStack(result.Error)
	}
	if result.RowsAffected > 0 && last.Email == email &&
		last.CreatedAt.Add(VerificationResendInterval).After(now) {
		return "", ErrVerificationTooSoon
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{"email", "hashed_secret", "expired_at", "created_at"}),
	}).Create(&Verification{
		Username:     username,
		Email:        email,
		HashedSecret: hashedSecret,
		ExpiredAt:    now.Add(VerificationLifetime),
		CreatedAt:    now,
	}).Error; err != nil {
		return "", errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return "", errors.WithStack(err)
	}

	return secret, nil
}

// ConfirmVerification 은 확인 링크의 비밀을 검증하고 그 주소를 확인된 것으로 표시한다.
// 없는 계정, 만료, 비밀 불일치, 그 사이에 바뀐 주소를 모두 ErrUnauthorized 하나로 답한다.
func (m *Manager) ConfirmVerification(username string, unhashedSecret string) (*User, error) {
	verification := Verification{}
	if err := m.db.Where(Verification{Username: username}).Take(&verification).Error; err != nil {
		return nil, ErrUnauthorized
	}

	// 만료된 행은 다시 쓸 일이 없다. 본 김에 치운다.
	if verification.ExpiredAt.Before(time.Now()) {
		if err := m.RevokeVerification(username); err != nil {
			logrus.Warnf("fail to remove expired verification of %s: %+v", username, err)
		}
		return nil, ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword(verification.HashedSecret, []byte(unhashedSecret)); err != nil {
		return nil, ErrUnauthorized
	}

	user, err := m.GetUser(username)
	if err != nil {
		return nil, ErrUnauthorized
	}

	// 링크가 날아간 뒤 주소를 또 바꿨으면 이 링크는 지금 주소를 증명하지 못한다.
	if user.Email != verification.Email {
		return nil, ErrUnauthorized
	}

	now := time.Now()
	if err := m.db.Model(&User{}).Where(User{Name: username}).
		Update("email_verified_at", now).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	user.EmailVerifiedAt = &now

	if err := m.RevokeVerification(username); err != nil {
		return nil, err
	}

	return user, nil
}

// RevokeVerification 은 확인 비밀을 폐기한다. 없어도 에러가 아니다.
func (m *Manager) RevokeVerification(username string) error {
	if err := m.db.Where(Verification{Username: username}).Delete(Verification{}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// newVerificationSecret 은 링크에 담을 비밀을 만든다.
// util.RandomString 은 math/rand 라 예측할 수 있어서 쓰지 않는다.
func newVerificationSecret() (string, error) {
	buf := make([]byte, verificationSecretSize)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.WithStack(err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
