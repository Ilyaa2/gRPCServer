package domain

import "errors"

var (
	ViolatedJsonContract = errors.New("the json contract between application and external server is violated")
)
