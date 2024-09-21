package ext

import "errors"

var ErrExtAlreadyExists = errors.New("extension already existing")
var ErrExtNotFound = errors.New("extension not found")
