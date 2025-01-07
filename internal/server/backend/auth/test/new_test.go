package test

import (
	"fmt"
	"runtime"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/sirupsen/logrus"
)

func newManager() (*auth.Manager, error) {
	logrus.SetLevel(logrus.TraceLevel)

	logrus.SetFormatter(&logrus.TextFormatter{DisableQuote: true, CallerPrettyfier: func(f *runtime.Frame) (string, string) {
		/* https://github.com/sirupsen/logrus/issues/63#issuecomment-476486166 */
		return "", fmt.Sprintf("%s:%d", f.File, f.Line)
	}})
	logrus.SetReportCaller(true)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	m, err := auth.New(db, "")
	if err != nil {
		return nil, err
	}
	return m, nil
}
