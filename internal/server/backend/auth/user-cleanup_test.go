package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

// countRows 는 username 에 딸린 행이 몇 개 남았는지 센다
func countRows(t *testing.T, db *gorm.DB, username string) map[string]int64 {
	t.Helper()

	counts := map[string]int64{}
	for table, where := range map[string]string{
		"tokens":        "username = ?",
		"recoveries":    "username = ?",
		"verifications": "username = ?",
		"members":       "user_name = ?",
	} {
		count := int64(0)
		require.NoError(t, db.Table(table).Where(where, username).Count(&count).Error)
		counts[table] = count
	}

	count := int64(0)
	require.NoError(t, db.Table("assigns").
		Where("subject_kind = ? AND subject_name = ?", auth.KindUser, username).
		Count(&count).Error)
	counts["assigns"] = count

	return counts
}

// 계정을 지우면 딸린 행이 전부 사라진다
func TestDeleteUserRemovesEverything(t *testing.T) {
	m, db, err := newManagerWithDB()
	require.NoError(t, err)

	_, _, err = m.Register("alice", "password",
		auth.WithEmail("alice@example.com"), auth.WithGroup("user"))
	require.NoError(t, err)

	// 지워야 할 것을 하나씩 만들어 둔다
	_, _, err = m.GenerateAccessKey("alice")
	require.NoError(t, err)

	_, err = m.IssueRecovery("alice")
	require.NoError(t, err)

	_, err = m.IssueVerification("alice", "alice@example.com")
	require.NoError(t, err)

	_, err = m.CreateRole("reader", []auth.Rule{})
	require.NoError(t, err)
	require.NoError(t, m.AssignRole(auth.Subject{Kind: auth.KindUser, Name: "alice"}, "reader"))

	before := countRows(t, db, "alice")
	for table, count := range before {
		require.NotZero(t, count, "%s should have a row before delete", table)
	}

	require.NoError(t, m.DeleteUser("alice"))

	for table, count := range countRows(t, db, "alice") {
		assert.Zero(t, count, "%s should be empty after delete", table)
	}
}

// 다른 사용자의 행은 건드리지 않는다
func TestDeleteUserKeepsOtherUsers(t *testing.T) {
	m, db, err := newManagerWithDB()
	require.NoError(t, err)

	_, err = m.CreateRole("reader", []auth.Rule{})
	require.NoError(t, err)

	for _, name := range []string{"alice", "bob"} {
		_, _, err = m.Register(name, "password",
			auth.WithEmail(name+"@example.com"), auth.WithGroup("user"))
		require.NoError(t, err)

		_, err = m.IssueRecovery(name)
		require.NoError(t, err)

		_, err = m.IssueVerification(name, name+"@example.com")
		require.NoError(t, err)

		require.NoError(t, m.AssignRole(auth.Subject{Kind: auth.KindUser, Name: name}, "reader"))
	}

	require.NoError(t, m.DeleteUser("alice"))

	for table, count := range countRows(t, db, "bob") {
		assert.NotZero(t, count, "%s of bob should survive", table)
	}
}

// 지운 이름으로 다시 가입해도 옛 비밀번호로는 들어올 수 없다
func TestDeleteUserThenRegisterAgain(t *testing.T) {
	m, err := newManager()
	require.NoError(t, err)

	_, _, err = m.Register("alice", "old-password", auth.WithGroup("user"))
	require.NoError(t, err)

	require.NoError(t, m.DeleteUser("alice"))

	_, _, err = m.Register("alice", "new-password", auth.WithGroup("user"))
	require.NoError(t, err)

	// 옛 비밀번호의 토큰이 남아 있었다면 여기서 통과해 버린다
	_, err = m.Default("alice", "old-password")
	assert.Error(t, err)

	_, err = m.Default("alice", "new-password")
	assert.NoError(t, err)
}

// 없는 계정을 지우면 이벤트도 나가지 않는다
func TestDeleteUserNotFound(t *testing.T) {
	m, _, events, err := newManagerWithHub()
	require.NoError(t, err)

	done := make(chan struct{})
	defer close(done)
	ch := events.Watch(auth.EventUserDeleted{}, done)

	assert.Error(t, m.DeleteUser("nobody"))

	select {
	case evt := <-ch:
		t.Fatalf("event should not be published: %v", evt)
	case <-time.After(100 * time.Millisecond):
	}
}

// 바깥에서도 삭제 이벤트를 받을 수 있다. 다른 백엔드가 붙을 자리다
func TestDeleteUserPublishesEvent(t *testing.T) {
	m, _, events, err := newManagerWithHub()
	require.NoError(t, err)

	_, _, err = m.Register("alice", "password", auth.WithEmail("alice@example.com"))
	require.NoError(t, err)

	received := make(chan auth.EventUserDeleted, 1)
	events.AddHandler(auth.EventUserDeleted{}, &recordDeleted{ch: received})

	require.NoError(t, m.DeleteUser("alice"))

	select {
	case got := <-received:
		assert.Equal(t, "alice", got.User.Name)
		// 지워진 User 를 통째로 담으므로 이름 말고도 들여다볼 수 있다
		assert.Equal(t, "alice@example.com", got.User.Email)
	case <-time.After(time.Second):
		t.Fatal("event was not delivered")
	}
}

// 구독자가 실패하면 DeleteUser 가 에러를 돌려준다. 계정 삭제 자체는 이미 끝나 있다
func TestDeleteUserReportsHandlerFailure(t *testing.T) {
	m, _, events, err := newManagerWithHub()
	require.NoError(t, err)

	_, _, err = m.Register("alice", "password")
	require.NoError(t, err)

	events.AddHandler(auth.EventUserDeleted{}, &failHandler{})

	err = m.DeleteUser("alice")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler is broken")

	// 에러가 났어도 계정은 없어졌다
	_, err = m.GetUser("alice")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

type recordDeleted struct {
	ch chan<- auth.EventUserDeleted
}

func (h *recordDeleted) Handle(ctx context.Context, evt pubsub.Event[any]) error {
	h.ch <- evt.Detail.(auth.EventUserDeleted)
	return nil
}

type failHandler struct{}

func (h *failHandler) Handle(ctx context.Context, evt pubsub.Event[any]) error {
	return errors.New("handler is broken")
}
