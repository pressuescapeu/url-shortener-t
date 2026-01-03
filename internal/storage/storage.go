package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrUrlExists     = errors.New("url exists")
	ErrNoURLDeleted  = errors.New("no url deleted")
	ErrInvalidAlias  = errors.New("invalid alias")
	ErrDatabaseError = errors.New("database error")
)
