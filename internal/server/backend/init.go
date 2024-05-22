package backend

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth/store"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "embed"
)

//go:embed policy.yaml
var defaultPolicy string

func initAuth(db *gorm.DB, salt string, initUser map[string]string) (*auth.Manager, error) {
	logrus.Trace("default policy: \n", defaultPolicy)
	s, err := store.Gorm(db, salt)
	if err != nil {
		return nil, err
	}

	m, err := auth.New(s, salt)
	if err != nil {
		return nil, err
	}

	for name, key := range initUser {
		logrus.Tracef("init user: %s %s", name, key)
		if _, _, err := m.Register(name, key, auth.WithGroup("admin", "user")); err != nil {
			logrus.Warn(err)
		}
	}

	return m, nil
}
func initDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)

	// TODO auto migrate
	if err := db.AutoMigrate(); err != nil {
		return nil, err
	}
	return db, err
}
