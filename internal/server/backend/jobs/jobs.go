package jobs

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub/v2"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

type Config struct {
}
type Manager struct {
	db             *gorm.DB
	events         *pubsub.Hub
	processContext context.Context
}

func New(ctx context.Context, conf *Config, db *gorm.DB, events *pubsub.Hub) (*Manager, error) {
	if err := db.AutoMigrate(
		&Job{},
	); err != nil {
		return nil, errors.WithStack(err)
	}
	return &Manager{db, events, ctx}, nil
}

type Job struct {
	Id         string     `json:"id" gorm:"primaryKey"`
	Name       string     `json:"name"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	Error      *string    `json:"error"`
	// QUESTION need log?
	// TODO tags map[string]struct{}
}

func (m *Manager) Run(ctx context.Context, name string, fn func(ctx context.Context) error) (*Job, error) {
	tx := m.db.WithContext(ctx).Model(Job{})

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		tx = tx.Debug()
	}

	job := Job{
		Id:        xid.New().String(),
		Name:      name,
		StartedAt: time.Now(),
	}

	if err := tx.Where(Job{Id: job.Id}).Save(&job).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	// run job on Background
	go runJob(m.processContext, m.db, job, fn)

	return &job, nil
}
func runJob(ctx context.Context, db *gorm.DB, job Job, fn func(ctx context.Context) error) {
	err := fn(ctx)
	if err != nil {
		errMsg := err.Error()
		job.Error = &errMsg
	}
	now := time.Now()
	job.FinishedAt = &now

	if err := db.WithContext(ctx).Save(&job).Error; err != nil {
		logrus.Error(err)
	}
}

func (m *Manager) Find(ctx context.Context, opts ...meta.ListOptionFn) (*meta.List[Job], error) {
	opt := meta.ListOption{
		Limit: 100,
	}

	for _, fn := range opts {
		fn(&opt)
	}

	return m.FindWithOption(ctx, opt)
}
func (m *Manager) FindWithOption(ctx context.Context, opt meta.ListOption) (*meta.List[Job], error) {
	if opt.Limit == 0 {
		opt.Limit = 20 // default value
	}

	list := meta.List[Job]{
		ListOption: opt,
	}

	tx := m.db.WithContext(ctx).Model(Job{})

	// TODO handle query

	if err := tx.Count(&list.Total).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	if err := tx.Limit(opt.Limit).Offset(opt.Offset).Find(&list.Items).Error; err != nil {
		return nil, err
	}

	return &list, nil
}
