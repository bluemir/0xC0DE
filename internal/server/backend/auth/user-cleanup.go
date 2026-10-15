package auth

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/bluemir/0xC0DE/internal/pubsub"
)

// EventUserDeleted 는 계정이 지워진 뒤에 발행된다.
// 이 시점에 users 행은 이미 없다. 딸린 데이터를 치우는 것은 받는 쪽의 몫이다.
//
// 지워진 User 를 통째로 담는다. 이름만 담으면 받는 쪽이 더 알아볼 방법이 없다.
type EventUserDeleted struct {
	User User
}

// userCleanup 은 계정이 지워졌을 때 auth 가 가진 행을 치운다.
//
// 포인터로 등록한다. pubsub 의 핸들러 집합이 sync.Map 이라 키가 해시 가능해야 하는데,
// Manager 를 값으로 품은 구조체는 해시할 수 없다.
type userCleanup struct {
	manager *Manager
}

func (h *userCleanup) Handle(ctx context.Context, evt pubsub.Event[any]) error {
	deleted, ok := evt.Detail.(EventUserDeleted)
	if !ok {
		return errors.Errorf("unexpected event detail for %T: %T", EventUserDeleted{}, evt.Detail)
	}

	return h.manager.deleteUserRows(deleted.User.Name)
}

// deleteUserRows 는 username 에 딸린 행을 한 트랜잭션에서 지운다.
//
// users 행을 지워도 이것들은 남는다. 특히 tokens 가 남으면 같은 이름으로 다시 가입했을 때
// 죽은 계정의 비밀번호와 passkey 가 새 계정에 그대로 붙는다.
func (m *Manager) deleteUserRows(username string) error {
	tx := m.db.Begin()
	defer tx.Rollback()

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

	return errors.WithStack(tx.Commit().Error)
}
