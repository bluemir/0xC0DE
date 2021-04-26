package store

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type Object interface {
	GetMetadata() *Metadata
}

type ObjectContainer struct {
	Object
}

func (obj *ObjectContainer) Scan(src interface{}) error {
	//logrus.Tracef("src type: %T", src)
	switch str := src.(type) {
	case []byte:
		if err := json.Unmarshal(str, obj); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(str), obj); err != nil {
			return err
		}
	default:
		return errors.Errorf("must []byte was '%T'", src)
	}

	return nil
}
func (obj ObjectContainer) Value() (driver.Value, error) {
	return json.Marshal(obj)
}
