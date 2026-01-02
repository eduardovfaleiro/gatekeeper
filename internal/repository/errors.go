package repository

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrUniqueConstraint = errors.New("unique constraint")
)
