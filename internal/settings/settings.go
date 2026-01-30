package settings

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"gorm.io/gorm"
)

type Config struct {
	Key    string `gorm:"primary_key"`
	At     time.Time
	Config []byte
}

type Manager struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// Load loads the configuration into the target struct.
// If the configuration does not exist, it initializes it using the initFunc.
// If initFunc returns nil for the config detail, it might skip saving or default handling depending on logic.
// Here we assume initFunc returns the initial config object.
func (m *Manager) Load(ctx context.Context, key string, target any, initFunc func() (any, error)) error {
	if err := m.db.AutoMigrate(&Config{}); err != nil {
		return errors.WithStack(err)
	}

	setting := Config{
		Key: key,
	}

	// Try to find the setting
	err := m.db.WithContext(ctx).First(&setting, "key = ?", key).Error
	if err == nil {
		// Found, unmarshal it
		if err := json.Unmarshal(setting.Config, target); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.WithStack(err)
	}

	// Not found, need initialization
	initConfig, err := initFunc()
	if err != nil {
		return err
	}

	// Serialize initial config
	buf, err := json.Marshal(initConfig)
	if err != nil {
		return errors.WithStack(err)
	}

	setting.At = time.Now()
	setting.Config = buf

	// Create new record
	if err := m.db.WithContext(ctx).Create(&setting).Error; err != nil {
		return errors.WithStack(err)
	}

	// Copy back to target
	// Re-marshalling is one way, or we can just copy if target is pointer to the same type as initConfig.
	// But initConfig is `any`. The safest way to ensure target is populated correctly
	// (especially if initConfig was just a struct value, not pointer) is to unmarshal the bytes we just made.
	if err := json.Unmarshal(buf, target); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (m *Manager) Save(ctx context.Context, key string, config any) error {
	buf, err := json.Marshal(config)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := m.db.WithContext(ctx).Save(&Config{
		Key:    key,
		At:     time.Now(),
		Config: buf,
	}).Error; err != nil {
		return errors.WithStack(err)
	}
	return nil
}
