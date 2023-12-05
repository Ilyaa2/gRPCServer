package domain

import "errors"

var (
	ErrExternalServer       = errors.New("external server error")
	ErrViolatedJsonContract = errors.New("the json contract between application and external server is violated" +
		" or there was incorrect data in the request")
	ErrIncorrectEmail  = errors.New("invalid Email")
	ErrInternalServer  = errors.New("internal server error")
	ErrNoInfoAvailable = errors.New("there was no information got from external server")
	ErrNilData         = errors.New("uninitialized data")

	ErrFileNotExists        = errors.New("the filepath is not correct")
	ErrInvalidFileStructure = errors.New("the structure of the file is defective")
	ErrIncorrectValue       = errors.New("the values are incorrect")
)
