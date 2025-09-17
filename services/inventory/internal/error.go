package internal

import "errors"

var (
	ErrItemNotFound         = errors.New("item not found")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrInsufficientReserved = errors.New("insufficient reserved stock")
)
