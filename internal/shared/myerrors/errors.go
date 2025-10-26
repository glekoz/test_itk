package myerrors

import "errors"

var (
	ErrNotFound       = errors.New("no result found")
	ErrInternal       = errors.New("something goes wrong")
	ErrAlreadyExists  = errors.New("already exists")
	ErrNegativeAmount = errors.New("amount can't be negative")
	ErrInvalidInput   = errors.New("only positive amount allowed")
)
