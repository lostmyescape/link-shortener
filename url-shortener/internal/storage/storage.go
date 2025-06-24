package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrURLExists     = errors.New("URL already exist")
	ErrAliasExists   = errors.New("alias already exists")
	ErrAliasNotFound = errors.New("alias not found")
)
