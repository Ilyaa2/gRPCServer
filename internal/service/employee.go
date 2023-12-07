package service

import (
	"context"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/repository"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	"gRPCServer/pkg/cache"
	"gRPCServer/pkg/util"
)

type EmployeeService struct {
	repo            repository.Employee
	reasons         *domain.AbsenceOptions
	compositeLogger domain.CompositeLogger
	cache           cache.Cache
}

func NewEmployeeService(repo repository.Employee, reasons *domain.AbsenceOptions,
	c cache.Cache, logger domain.CompositeLogger) *EmployeeService {
	return &EmployeeService{repo: repo,
		reasons:         reasons,
		cache:           c,
		compositeLogger: logger,
	}
}

func (e *EmployeeService) GetReasonOfAbsence(ctx context.Context,
	details *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	var response *dataModification.ContactDetails
	ok := false

	r, err := e.cache.Get(details.Email)
	if err == nil {
		response, ok = r.(*dataModification.ContactDetails)
	}

	if ok {
		e.compositeLogger.ApplicationLogger.Debug(
			"got contact details from cache",
			map[string]interface{}{
				"package":         "service",
				"function":        "GetReasonOfAbsence",
				"contact-details": response,
			},
		)
	}

	if !ok {
		empData, err := e.repo.GetByEmail(ctx, details.Email)
		e.compositeLogger.ApplicationLogger.Debug(
			"got data from GetByEmail repo's method",
			map[string]interface{}{
				"package":       "service",
				"function":      "GetReasonOfAbsence",
				"employee-data": empData,
				"err":           err,
			},
		)
		if err != nil {
			return nil, err
		}

		if empData.Status != "OK" || empData.Data == nil {
			return nil, domain.ErrNoInfoAvailable
		}

		absReason, err := e.repo.GetAbsenceReason(ctx, empData)
		e.compositeLogger.ApplicationLogger.Debug(
			"got data from GetAbsenceReason repo's method",
			map[string]interface{}{
				"package":        "service",
				"function":       "GetReasonOfAbsence",
				"absence-reason": absReason,
				"err":            err,
			},
		)
		if err != nil {
			return nil, err
		}

		if absReason.Status != "OK" || absReason.Data == nil {
			return nil, domain.ErrNoInfoAvailable
		}

		oneAbsReason := absReason.Data[0]
		oneEmpData := empData.Data[0]

		reason, okay := e.reasons.GetReason(oneAbsReason.ReasonId)

		response = &dataModification.ContactDetails{
			DisplayName: oneEmpData.DisplayName,
			Email:       oneEmpData.Email,
			WorkPhone:   oneEmpData.WorkPhone,
		}
		if !okay {
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
			e.cache.Set(details.Email, response, 0)
		}
	}
	e.compositeLogger.ApplicationLogger.Debug(
		"returning data",
		map[string]interface{}{
			"package":  "service",
			"function": "GetReasonOfAbsence",
			"response": response,
		},
	)
	return response, nil
}
