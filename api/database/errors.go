package database

import "errors"

var (
	ErrorNotFound = errors.New("not found")
	ErrorConflict = errors.New("conflicted")
)
