package store

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// assumption
//   all obj must have primary key
//   list operation must full scan, just filtering

// orm but based on etcd
type IStore interface {
	Create(ctx context.Context, obj any) error
	Load(ctx context.Context, obj any) error
	List(ctx context.Context, t any, condition func(any) bool) ([]any, error)
	Stream(ctx context.Context, t any, condition func(any) bool) (<-chan any, error)
	Update(ctx context.Context, obj any) error
	Save(ctx context.Context, obj any) error
	Delete(ctx context.Context, obj any) error
}

var _ IStore = &Store{}

func New(ctx context.Context, endpoint string, opts ...OptionFn) (*Store, error) {
	opt := &Option{
		Endpoint: endpoint,
	}

	for _, fn := range opts {
		fn(opt)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		cli.Close()
	}()

	return &Store{cli}, nil
}

type OptionFn func(*Option)

type Option struct {
	Endpoint string
	TLS      struct {
		Cert struct {
			Cert string
			Key  string
		}
		CA string
	}
}
type Store struct {
	client *clientv3.Client
}

var ErrNotImplements = errors.New("not implements")

func (s *Store) Load(ctx context.Context, obj any) error {
	return ErrNotImplements
}
func (s *Store) Create(ctx context.Context, obj any) error {
	return ErrNotImplements
}
func (s *Store) List(ctx context.Context, t any, condition func(any) bool) ([]any, error) {
	return nil, ErrNotImplements
}
func (s *Store) Stream(ctx context.Context, t any, condition func(any) bool) (<-chan any, error) {
	return nil, ErrNotImplements
}
func (s *Store) Update(ctx context.Context, obj any) error {
	return ErrNotImplements
}
func (s *Store) Save(ctx context.Context, obj any) error {
	return ErrNotImplements
}
func (s *Store) Delete(ctx context.Context, obj any) error {
	return ErrNotImplements
}

func GetTypeString(obj any) string {
	// TODO https://github.com/go-gorm/gorm/blob/master/schema/schema.go#L121-L145
	tname := reflect.TypeOf(obj).String()
	name := TakeLastName(tname)
	return CamelCaseToKebabCase(name)
}

func TakeLastName(str string) string {
	arr := strings.Split(str, ".")
	return arr[len(arr)-1]
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func CamelCaseToKebabCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}
