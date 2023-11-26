package mock_repository

import (
	"context"
	"gRPCServer/internal/domain"
	"time"
)

type EmployeeRepoMock struct {
}

func (e *EmployeeRepoMock) GetByEmail(_ context.Context, _ string) (*domain.EmployeeData, error) {
	respData := &domain.EmployeeData{
		Status: "OK",
		Data: []domain.InnerData{
			{
				Id:          1234,
				DisplayName: "Changed Name",
				Email:       "example@gmail.com",
				WorkPhone:   "!",
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
