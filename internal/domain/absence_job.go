package domain

import (
	"context"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
)

// AbsenceJob is a job that related to request with absence reason
type AbsenceJob struct {
	Context context.Context
	Input   *dataModification.ContactDetails
	Result  chan Future
}

// Future is a structure that is used by the handler to block on it and
// wait for a response. And the worker puts the answer there as soon as he calculates the request.
type Future struct {
	Error  error
	Output *dataModification.ContactDetails
}
