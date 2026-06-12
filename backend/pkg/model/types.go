package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Ingredient 食材条目（百科 / 版本共用）
type Ingredient struct {
	Name   string   `json:"name"`
	Amount float64  `json:"amount"`
	Unit   string   `json:"unit"`
	Note   string   `json:"note,omitempty"`
}

// ProcessStep 制作步骤
type ProcessStep struct {
	Order           int     `json:"order"`
	Content         string  `json:"content"`
	DurationMinutes *int    `json:"duration_minutes,omitempty"`
	ImageURL        *string `json:"image_url,omitempty"`
}

// JSONSlice 泛型 JSON 数组，用于 GORM 自定义类型
type JSONSlice[T any] []T

func (j JSONSlice[T]) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONSlice[T]) Scan(value interface{}) error {
	if value == nil {
		*j = JSONSlice[T]{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONSlice: unsupported type %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// StringSlice JSON 字符串数组（如 tags、images）
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("StringSlice: unsupported type %T", value)
	}
	return json.Unmarshal(bytes, s)
}
