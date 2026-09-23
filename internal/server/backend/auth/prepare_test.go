package auth_test

import (
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
)

func newManager() (*auth.Manager, error) {
	m, _, err := newManagerWithDB()
	return m, err
}

// newManagerWithDB 는 db 도 같이 돌려준다.
// 만료나 재발송 간격처럼 시각에 걸린 동작을 시험하려면 저장된 값을 직접 되돌려야 한다.
func newManagerWithDB() (*auth.Manager, *gorm.DB, error) {
	logrus.SetLevel(logrus.TraceLevel)

	logrus.SetFormatter(&logrus.TextFormatter{DisableQuote: true, CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		/* https://github.com/sirupsen/logrus/issues/63#issuecomment-476486166 */
		return "", fmt.Sprintf("%s:%d", f.File, f.Line)
	}})
	logrus.SetReportCaller(true)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	m, err := auth.New(db, &auth.Config{})
	if err != nil {
		return nil, nil, err
	}
	return m, db, nil
}
