package controller

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/bluemir/0xC0DE/internal/server/backend"
)

// Controller watches job lifecycle events and performs post-processing.
type Controller struct {
	ctx      context.Context
	backends *backend.Backends
}

// New creates a Controller. Call Run() to start the event loop.
func New(ctx context.Context, bs *backend.Backends) *Controller {
	return &Controller{ctx: ctx, backends: bs}
}

// Run starts the event loop. It blocks until the context is cancelled.
func (c *Controller) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	//eg.Go(c.handleJob(ctx))

	return eg.Wait()
}
