package domain

import "errors"

var (
	ErrViolatedJsonContract = errors.New("the json contract between application and external server is violated")
	ErrInvalidEmail         = errors.New("invalid Email")
)
