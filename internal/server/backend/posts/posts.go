package posts

import (
	"time"

	"github.com/rs/xid"
	"golang.org/x/net/context"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub/v1"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

type Config struct {
}
type Manager struct {
	db     *gorm.DB
	events *pubsub.Hub
}

func New(ctx context.Context, conf *Config, db *gorm.DB, events *pubsub.Hub) (*Manager, error) {
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

func (m *Manager) Create(ctx context.Context, message string) (*Post, error) {
	post := &Post{
		Id:      xid.New().String(),
		At:      time.Now(),
		Message: message,
	}
	if err := m.db.Create(post).Error; err != nil {
		return nil, err
	}

	m.events.Publish("posts.created", post)

	return post, nil
}

func (m *Manager) List(ctx context.Context, opts ...meta.ListOptionFn) ([]Post, error) {
	opt := meta.ListOption{
		Limit: 20,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.ListWithOption(ctx, &opt)
}
func (m *Manager) ListWithOption(ctx context.Context, opt *meta.ListOption) ([]Post, error) {
	if opt.Limit == 0 {
		opt.Limit = 20
	}

	posts := []Post{}

	if err := m.db.Limit(opt.Limit).Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}
