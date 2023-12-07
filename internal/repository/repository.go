package repository

import (
	"context"
	"gRPCServer/internal/domain"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

//This file is needed to define the abstraction of all repositories in my project.

// Employee is the lowest layer of my application.
// It works directly with data and making the appropriate requests to get it.
type Employee interface {
	GetByEmail(ctx context.Context, email string) (*domain.EmployeeData, error)
	GetAbsenceReason(ctx context.Context, empData *domain.EmployeeData) (*domain.AbsenceReason, error)
}
