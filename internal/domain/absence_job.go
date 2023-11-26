package domain

import (
	"gRPCServer/internal/transport/grpc/sources/dataModification"
)

type AbsenceJob struct {
	Input  *dataModification.ContactDetails
	Result chan Future
}

type Future struct {
	Error  error
	Output *dataModification.ContactDetails
}
