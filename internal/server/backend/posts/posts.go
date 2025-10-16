package posts

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"
	"golang.org/x/net/context"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub/v2"
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
type EventPostCreated struct {
	Post Post
}

func (m *Manager) Create(ctx context.Context, message string) (*Post, error) {
	post := &Post{
		Id:      xid.New().String(),
		At:      time.Now(),
		Message: message,
	}
	if err := m.db.WithContext(ctx).Create(post).Error; err != nil {
		return nil, err
	}

	m.events.Publish(ctx, EventPostCreated{Post: *post})

	return post, nil
}

func (m *Manager) List(ctx context.Context, opts ...meta.ListOptionFn) (*meta.List[Post], error) {
	opt := meta.ListOption{
		Limit: 20,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.ListWithOption(ctx, &opt)
}
func (m *Manager) ListWithOption(ctx context.Context, opt *meta.ListOption) (*meta.List[Post], error) {
	list := meta.List[Post]{}

	if err := m.db.WithContext(ctx).Limit(opt.Limit).Find(&list.Items).Error; err != nil {
		return nil, err
	}

	return &list, nil
}

type Query struct {
	Message *string
}

func (m *Manager) FindWithOption(ctx context.Context, query Query, opt *meta.ListOption) (*meta.List[Post], error) {
	tx := m.db.WithContext(ctx)

	if query.Message != nil && len(*query.Message) > 0 {
		tx = tx.Where("message LIKE @message", map[string]any{
			"message": "%" + *query.Message + "%",
		})
	}

	list := meta.List[Post]{}

	if err := tx.Limit(opt.Limit).Offset(opt.Offset).Find(&list.Items).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &list, nil
}
