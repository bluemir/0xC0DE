package auth

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// TODO  gob, base64
func (set *Set) Scan(src interface{}) error {
	//logrus.Tracef("src type: %T", src)
	switch str := src.(type) {
	case []byte:
		if err := json.Unmarshal(str, set); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(str), set); err != nil {
			return err
		}
	default:
		return errors.Errorf("must []byte was '%T'", src)
	}

	return nil
}
func (set Set) Value() (driver.Value, error) {
	return json.Marshal(set)
}

func (labels *Labels) Scan(src interface{}) error {
	//logrus.Tracef("src type: %T", src)
	switch str := src.(type) {
	case []byte:
		if err := json.Unmarshal(str, labels); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(str), labels); err != nil {
			return err
		}
	default:
		return errors.Errorf("must []byte was '%T'", src)
	}

	return nil
}
func (labels Labels) Value() (driver.Value, error) {
	return json.Marshal(labels)
}

func (kvs *KeyValues) Value() (driver.Value, error) {
	return json.Marshal(kvs)
}
func (kvs *KeyValues) Scan(src interface{}) error {
	switch str := src.(type) {
	case []byte:
		if err := json.Unmarshal(str, kvs); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(str), kvs); err != nil {
			return err
		}
	default:
		return errors.Errorf("must []byte was '%T'", src)
	}

	return nil
}
