package auth_test

import (
	"context"
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func newManager() (*auth.Manager, error) {
	m, _, err := newManagerWithDB()
	return m, err
}

// newManagerWithHub 는 hub 까지 돌려준다.
// 계정 삭제 이벤트처럼 Manager 밖으로 나가는 동작을 시험할 때 쓴다.
func newManagerWithHub() (*auth.Manager, *gorm.DB, *pubsub.Hub, error) {
	return newManagerWith(context.Background())
}

// newManagerWithDB 는 db 도 같이 돌려준다.
// 만료나 재발송 간격처럼 시각에 걸린 동작을 시험하려면 저장된 값을 직접 되돌려야 한다.
func newManagerWithDB() (*auth.Manager, *gorm.DB, error) {
	m, db, _, err := newManagerWith(context.Background())
	return m, db, err
}

func newManagerWith(ctx context.Context) (*auth.Manager, *gorm.DB, *pubsub.Hub, error) {
	logrus.SetLevel(logrus.TraceLevel)

	logrus.SetFormatter(&logrus.TextFormatter{DisableQuote: true, CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		/* https://github.com/sirupsen/logrus/issues/63#issuecomment-476486166 */
		return "", fmt.Sprintf("%s:%d", f.File, f.Line)
	}})
	logrus.SetReportCaller(true)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	events, err := pubsub.NewHub(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	m, err := auth.New(ctx, &auth.Config{}, db, events)
	if err != nil {
		return nil, nil, nil, err
	}
	return m, db, events, nil
}
