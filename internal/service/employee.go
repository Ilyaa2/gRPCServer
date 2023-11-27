package service

import (
	"context"
	"errors"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/repository"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
)

type EmployeeService struct {
	repo    repository.Employee
	reasons *domain.AbsenceOptions
}

func NewEmployeeService(repo repository.Employee, reasons *domain.AbsenceOptions) *EmployeeService {
	return &EmployeeService{repo: repo, reasons: reasons}
}

// todo нужно создать типы ошибок
func (e *EmployeeService) GetReasonOfAbsence(ctx context.Context, details *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	empData, err := e.repo.GetByEmail(ctx, details.Email)
	if err != nil {
		return nil, err
	}

	//todo добавил только один id
	//TODO ТОЛЬКО ОДИН ID - ПЕРВЫЙ.
	absReason, err := e.repo.GetAbsenceReason(ctx, empData)
	if err != nil {
		return nil, err
	}
	oneAbsReason := absReason.Data[0]
	oneEmpData := empData.Data[0]

	reason, ok := e.reasons.GetReason(oneAbsReason.ReasonId)
	if !ok {
		return nil, errors.New("the reason id isn't appropriate to the document given")
	}

	return &dataModification.ContactDetails{
		DisplayName: oneEmpData.DisplayName + " " + reason.Emoji + " " + reason.Description,
		Email:       oneEmpData.Email,
		WorkPhone:   oneEmpData.WorkPhone,
	}, nil
}
