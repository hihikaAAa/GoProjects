package storage

import(
	"errors"
)

var (
	errURLNotFound = errors.New("url not found")
	ErrUrlExists = errors.New("url exists")
)