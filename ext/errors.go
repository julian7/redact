package ext

import "errors"

var ErrExtAlreadyExists = errors.New("extension already exists")
var ErrExtNotFound = errors.New("extension not found")
