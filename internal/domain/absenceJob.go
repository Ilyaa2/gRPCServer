package domain

import "gRPCServer/internal/transport/grpc/sources/dataModification"

type AbsenceJob struct {
	Data   *dataModification.ContactDetails
	Result chan string
}
