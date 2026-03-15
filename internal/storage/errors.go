package storage

import "errors"

var (
	ErrNotFound        = errors.New("swallow: record not found")
	ErrDuplicate       = errors.New("swallow: record already exists")
	ErrNoActiveProject = errors.New("swallow: no active project set")
)
