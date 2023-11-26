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
// todo ФУНДАМЕНТАЛЬНО НЕПРАВИЛЬНО должен возвращать измененный dm.ContactDetails. В ФИО дописать через словарь причину отсутствия. Нужно ли блокироваться при поиске в словаре?
// todo Скорее всего нужно будет реализовать свою Unmodifiable/Immutable map - просто свою структуру.

func (e *EmployeeService) GetReasonOfAbsence(ctx context.Context, details *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	empData, err := e.repo.GetByEmail(ctx, details.Email)
	if err != nil {
		return nil, err
	}
	//todo добавил только один id
	absReason, err := e.repo.GetAbsenceReason(ctx, empData)
	//TODO воспользоваться reasons
	if err != nil {
		return nil, err
	}
	r := absReason.Data[0]

	reason, ok := e.reasons.GetReason(r.ReasonId)
	if !ok {
		return nil, errors.New("the reason id isn't appropriate to the document given")
	}

	return &dataModification.ContactDetails{
		DisplayName: details.DisplayName + reason.Emoji + reason.Description,
		Email:       details.Email,
		MobilePhone: details.MobilePhone,
		//TODO FIX THIS
		WorkPhone: details.WorkPhone + details.WorkPhone,
	}, nil
}
