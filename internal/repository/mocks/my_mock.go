package mock_repository

import (
	"context"
	"gRPCServer/internal/domain"
	"time"
)

type EmployeeRepoMock struct {
}

func (e *EmployeeRepoMock) GetByEmail(_ context.Context, email string) (*domain.EmployeeData, error) {
	respData := &domain.EmployeeData{
		Status: "OK",
		Data: []domain.EmployeeInnerData{
			{
				Id:          1234,
				DisplayName: "Иванов Семен Петрович",
				Email:       email,
				WorkPhone:   "",
			},
		},
	}

	return respData, nil
}

func (e *EmployeeRepoMock) GetAbsenceReason(_ context.Context, _ *domain.EmployeeData) (*domain.AbsenceReason, error) {
	respData := &domain.AbsenceReason{
		Status: "OK",
		Data: []domain.AbsenceReasonData{
			{
				Id:          1,
				PersonId:    1234,
				CreatedDate: time.DateTime,
				DateFrom:    time.DateTime,
				DateTo:      time.DateTime,
				ReasonId:    1,
			},
		},
	}

	return respData, nil
}
