package backend

import (
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth/store"
	"github.com/bluemir/0xC0DE/internal/server/middleware/auth/verb"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func initAuth(db *gorm.DB, salt string, initUser map[string]string) (*auth.Manager, error) {
	s, err := store.Gorm(db, salt)
	if err != nil {
		return nil, err
	}

	m, err := auth.New(s, salt)
	if err != nil {
		return nil, err
	}

	if _, err := m.CreateGroup("admin"); err != nil {
		return nil, err
	}

	for name, key := range initUser {
		logrus.Tracef("init user: %s %s", name, key)
		u, _, err := m.Register(name, key)
		if err != nil {
			return nil, err
		}
		u.Groups.Add("admin", "user")
		if err := m.UpdateUser(u); err != nil {
			return nil, err
		}
	}

	m.CreateRole("admin", []auth.Rule{
		{
			Resource: auth.KeyValues{
				"kind": "admin-page",
			},
		},
		{
			Resource: auth.KeyValues{
				"kind": "server",
			},
		},
		{
			Resource: auth.KeyValues{
				"kind": "user",
			},
		},
	})
	m.CreateRole("user", []auth.Rule{
		{
			Verbs: []auth.Verb{
				verb.Update,
			},
			Resource: auth.KeyValues{
				"kind": "user",
			},
			Conditions: []auth.Condition{
				`user.name == resource.name`,
			},
		},
	})

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
