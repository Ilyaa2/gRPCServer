package service

import (
	"context"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
)

//This file is needed to define the abstraction of all handlers in my project

//go:generate mockgen -source=service.go -destination=mocks/mock.go

// Employee this is the middle layer of the application.
// It issues commands based on the received request to the lower layer and intermediately processes the data.
type Employee interface {
	GetReasonOfAbsence(ctx context.Context, details *dataModification.ContactDetails) (*dataModification.ContactDetails, error)
}

type Services struct {
	Employee Employee
}
