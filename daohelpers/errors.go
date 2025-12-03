package daohelpers

import (
	"errors"
)

var (
	ErrDuplicate = errors.New("DAO action cannot be performed due to the conflicts on unique fields")
)
