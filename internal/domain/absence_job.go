package domain

import (
	"context"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
)

type AbsenceJob struct {
	Context context.Context
	Input   *dataModification.ContactDetails
	Result  chan Future
}

type Future struct {
	Error  error
	Output *dataModification.ContactDetails
}
