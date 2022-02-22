package store

import (
	"bytes"
	"encoding/gob"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func New(db *gorm.DB) (*Store, error) {
	if err := db.AutoMigrate(
		&Pair{},
		&Label{},
		&TempKV{},
	); err != nil {
		return nil, err
	}
	return &Store{db}, nil

}

type Store struct {
	db *gorm.DB
}

type Pair struct {
	Kind      string `gorm:"primary_key"`
	Id        string `gorm:"primary_key"`
	Namespace string `gorm:"primary_key"`
	Rev       int
	Value     []byte
}
type Label struct {
	Kind  string `gorm:"primary_key"`
	Id    string `gorm:"primary_key"`
	Key   string `gorm:"primary_key"`
	Value string
}
type TempKV struct {
	Key   string
	Value string
}

func (s *Store) Save(obj Object) error {
	logrus.Tracef("%#v", obj)

	meta := obj.GetMetadata()

	var err error
	txn := s.db.Begin()
	defer func() {
		if err != nil {
			txn.Rollback()
		}
	}()

	pair := &Pair{}

	err = txn.Where(&Pair{Kind: meta.Kind, Id: meta.Id, Namespace: meta.Namespace}).Take(pair).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			logrus.Tracef("recode not found")
		} else {
			txn.Rollback()
			return err
		}
	}

	if pair.Rev > meta.Rev {
		txn.Rollback()
		return errors.Errorf("conflict")
	}

	meta.Rev += 1

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(obj); err != nil {
		return err
	}

	err = txn.Save(&Pair{
		Kind:      meta.Kind,
		Id:        meta.Id,
		Namespace: meta.Namespace,
		Rev:       meta.Rev,
		Value:     buf.Bytes(),
	}).Error
	if err != nil {
		return err
	}

	for k, v := range meta.Labels {
		err = txn.Save(&Label{
			Key:   k,
			Value: v,
			Kind:  meta.Kind,
			Id:    meta.Id,
		}).Error
		if err != nil {
			return err
		}
	}

	return txn.Commit().Error
}
func (s *Store) Load(meta *Metadata, obj Object) error {
	logrus.Tracef("%#v", meta)

	if meta.Id == "" {
		return gorm.ErrRecordNotFound
	}
	if meta.Namespace == "" {
		return gorm.ErrRecordNotFound
	}
	pair := &Pair{}

	if err := s.db.Where(&Pair{
		Kind:      meta.Kind,
		Id:        meta.Id,
		Namespace: meta.Namespace,
	}).Take(pair).Error; err != nil {
		return err
	}

	if err := gob.NewDecoder(bytes.NewReader(pair.Value)).Decode(obj); err != nil {
		return err
	}
	return nil
}

func (s *Store) Find(kind string, namespace string, labels map[string]string) ([]Metadata, error) {
	logrus.Tracef("%s %s %#v", kind, namespace, labels)

	txn := s.db.Begin()
	defer txn.Rollback()

	for k, v := range labels {
		txn.Create(&TempKV{k, v})
	}

	result := []Metadata{}
	pairs := []Pair{}

	switch len(labels) {
	case 0:
		if err := txn.Table("pairs").
			Where(`namespace=? AND kind=?`, namespace, kind).
			Find(&pairs).Error; err != nil {
			return nil, err
		}
	default:
		if err := txn.Table("pairs").
			Joins("LEFT JOIN labels ON pairs.kind = labels.kind AND pairs.id = labels.id").
			Where(`(labels.key, labels.value) in (select * from temp_kvs) AND namespace=? AND pairs.kind=?`, namespace, kind).
			Group(`pairs.kind, pairs.id`).
			Having("count(*) = ?", len(labels)).
			Find(&pairs).Error; err != nil {
			return nil, err
		}
	}

	for _, pair := range pairs {
		result = append(result, Metadata{
			Id:        pair.Id,
			Kind:      pair.Kind,
			Rev:       pair.Rev,
			Namespace: pair.Namespace,
			Labels:    labels,
		})
	}

	return result, nil
}

func IsNotFoundError(err error) bool {
	return err == gorm.ErrRecordNotFound
}