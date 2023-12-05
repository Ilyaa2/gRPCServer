package service

import (
	"context"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/repository"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	"gRPCServer/pkg/util"
)

type EmployeeService struct {
	repo            repository.Employee
	reasons         *domain.AbsenceOptions
	compositeLogger domain.CompositeLogger
}

func NewEmployeeService(repo repository.Employee, reasons *domain.AbsenceOptions, logger domain.CompositeLogger) *EmployeeService {
	return &EmployeeService{repo: repo, reasons: reasons, compositeLogger: logger}
}

func (e *EmployeeService) GetReasonOfAbsence(ctx context.Context, details *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	empData, err := e.repo.GetByEmail(ctx, details.Email)
	if err != nil {
		return nil, err
	}

	if empData.Status != "OK" || empData.Data == nil {
		return nil, domain.ErrNoInfoAvailable
	}

	absReason, err := e.repo.GetAbsenceReason(ctx, empData)
	if err != nil {
		return nil, err
	}

	if absReason.Status != "OK" || absReason.Data == nil {
		return nil, domain.ErrNoInfoAvailable
	}

	oneAbsReason := absReason.Data[0]
	oneEmpData := empData.Data[0]

	reason, ok := e.reasons.GetReason(oneAbsReason.ReasonId)

	response := &dataModification.ContactDetails{
		DisplayName: oneEmpData.DisplayName,
		Email:       oneEmpData.Email,
		WorkPhone:   oneEmpData.WorkPhone,
	}
	if !ok {
		e.compositeLogger.ApplicationLogger.Warn(
			"didn't find the reason in the reasons dictionary",
			map[string]interface{}{
				"req-id":   util.GetReqIDFromContext(ctx),
				"reasonID": oneAbsReason.ReasonId,
			},
		)
	} else {
		if reason.Emoji != "" {
			response.DisplayName += " - " + reason.Emoji + " " + reason.Description
		} else {
			response.DisplayName += " - " + reason.Description
		}
	}

	return response, nil
}
