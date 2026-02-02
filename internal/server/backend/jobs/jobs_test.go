package jobs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bluemir/0xC0DE/internal/pubsub"
	"github.com/bluemir/0xC0DE/internal/server/backend/jobs"
	"github.com/bluemir/0xC0DE/internal/server/backend/meta"
)

func newTestManager(t *testing.T) *jobs.Manager {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	ctx := context.Background()
	hub, err := pubsub.NewHub(ctx)
	require.NoError(t, err)

	conf := &jobs.Config{}
	m, err := jobs.New(ctx, conf, db, hub)
	require.NoError(t, err)

	return m
}

func TestJobRunAndFind(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()

	done := make(chan struct{})

	job, err := m.Run(ctx, "test-job", func(ctx context.Context) error {
		close(done)
		return nil
	})
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "test-job", job.Name)
	assert.NotEmpty(t, job.Id)

	// Wait for job to finish
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for job")
	}

	// Verify job status (FinishedAt should be set)
	// We need to fetch it again from DB because runJob is async and updates the record.
	// But finding by ID requires m.Find to support ID filter or we check list.
	// jobs.go FindWithOption supports meta.ListOption. We can use that if we implement filtering or just list all.
	// jobs.go TODO handle query says "TODO handle query". So maybe ID filter not implemented in search yet?
	// But we can check DB directly using gorm if needed, or rely on List returning it.

	// Let's use List
	list, err := m.Find(ctx, meta.Limit(10))
	assert.NoError(t, err)
	assert.Len(t, list.Items, 1)

	foundJob := list.Items[0]
	assert.Equal(t, job.Id, foundJob.Id)

	// Since runJob runs in background and updates DB, we might need to retry check if it's lagging slightly
	// (though in-memory sqlite is fast, goroutine scheduling is non-deterministic).
	// We waited for 'done' channel, which is closed INSIDE the job function.
	// runJob calls fn(ctx) then updates DB. So 'done' closed means fn returned.
	// But updating DB happens matches AFTER fn returns.
	// So we need to wait a tiny bit more for DB update to complete.
	time.Sleep(100 * time.Millisecond)

	list, err = m.Find(ctx, meta.Limit(10))
	assert.NoError(t, err)
	foundJob = list.Items[0]

	assert.NotNil(t, foundJob.FinishedAt)
	assert.Nil(t, foundJob.Error)
}

func TestJobFailure(t *testing.T) {
	m := newTestManager(t)
	ctx := context.Background()
	done := make(chan struct{})

	_, err := m.Run(ctx, "fail-job", func(ctx context.Context) error {
		defer close(done)
		return errors.New("job failed")
	})
	assert.NoError(t, err)

	<-done
	time.Sleep(100 * time.Millisecond)

	list, err := m.Find(ctx, meta.Limit(1))
	assert.NoError(t, err)
	assert.Len(t, list.Items, 1)

	foundJob := list.Items[0]
	assert.NotNil(t, foundJob.FinishedAt)
	assert.NotNil(t, foundJob.Error)
	assert.Equal(t, "job failed", *foundJob.Error)
}
