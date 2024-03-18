package backend

import (
	"context"
	"os"

	"github.com/bluemir/0xC0DE/internal/events"
	"github.com/bluemir/0xC0DE/internal/server/backend/auth"
	"github.com/bluemir/0xC0DE/internal/server/backend/posts"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Commandline options
type Args struct {
	ConfigFilePath string
	DBPath         string
	Salt           string
	InitUser       map[string]string
}

func NewArgs() Args {
	return Args{
		InitUser: map[string]string{},
	}
}

// Config from file
type Config struct {
	Posts posts.Config
}
type Backends struct {
	Auth   *auth.Manager
	Events *events.Hub
	Posts  *posts.Manager
}

func Initialize(ctx context.Context, args *Args) (*Backends, error) {
	conf, err := readCofigFile(args.ConfigFilePath)
	if err != nil {
		return nil, err
	}
	events, err := events.NewHub(ctx)
	if err != nil {
		return nil, err
	}
	// init components
	db, err := initDB(args.DBPath)
	if err != nil {
		return nil, errors.Wrapf(err, "init server failed")
	}
	authManager, err := initAuth(db, args.Salt, args.InitUser)
	if err != nil {
		return nil, errors.Wrapf(err, "init server failed")
	}
	postManager, err := posts.New(ctx, &conf.Posts, db, events)
	if err != nil {
		return nil, errors.Wrapf(err, "init post manager failed")
	}

	return &Backends{
		Events: events,
		Auth:   authManager,
		Posts:  postManager,
	}, nil
}
func readCofigFile(configFilePath string) (*Config, error) {
	buf, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	conf := Config{}
	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
