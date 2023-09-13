package posts

import (
	"time"

	"github.com/rs/xid"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/bus"
)

type Manager struct {
	db     *gorm.DB
	events *bus.Bus
}

func New(db *gorm.DB, events *bus.Bus) (*Manager, error) {
	if err := db.AutoMigrate(&Post{}); err != nil {
		return nil, err
	}
	return &Manager{db, events}, nil
}

type Post struct {
	Id      string    `json:"id"`
	At      time.Time `json:"at"`
	Message string    `json:"message"`
}

func (m *Manager) Create(message string) (*Post, error) {
	post := &Post{
		Id:      xid.New().String(),
		At:      time.Now(),
		Message: message,
	}
	if err := m.db.Create(post).Error; err != nil {
		return nil, err
	}

	m.events.FireEvent("posts/created", post)

	return post, nil
}

type ListOption struct {
	Limit int
}

func (m *Manager) List(opt ListOption) ([]Post, error) {
	if opt.Limit == 0 {
		opt.Limit = 20
	}

	posts := []Post{}

	if err := m.db.Limit(opt.Limit).Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}
