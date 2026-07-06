package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StringList []string

func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	data, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = StringList{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("scan StringList: unsupported value %T", value)
	}
	var decoded []string
	if len(raw) == 0 {
		decoded = []string{}
	} else if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	*s = StringList(decoded)
	return nil
}

func (s StringList) Slice() []string {
	if s == nil {
		return []string{}
	}
	return []string(s)
}
