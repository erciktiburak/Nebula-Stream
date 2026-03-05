package state

import "errors"

var ErrNotFound = errors.New("state key not found")

type Store interface {
	Save(key string, data []byte) error
	Load(key string) ([]byte, error)
}
