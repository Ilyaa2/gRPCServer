package service

import (
	"context"
	"gRPCServer/internal/domain"
	mock_repository "gRPCServer/internal/repository/mocks"
	dm "gRPCServer/internal/transport/grpc/sources/dataModification"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

type mockBehavior func(*mock_repository.MockEmployee, string)

type test struct {
	testName     string
	input        *dm.ContactDetails
	mockBehavior mockBehavior
	wantData     *dm.ContactDetails
	wantErr      error
}

func initTests() *[]test {
	testTable := []test{
		{
			testName: "OK",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "example@gmail.com",
				WorkPhone:   "+71234567890",
			},
			mockBehavior: func(s *mock_repository.MockEmployee, email string) {
				empData := &domain.EmployeeData{
					Status: "OK",
					Data: []domain.EmployeeInnerData{
						{
							Id:          1234,
							DisplayName: "Иванов Семен Петрович",
							Email:       email,
							WorkPhone:   "+71234567890",
						},
					},
				}

				s.EXPECT().GetByEmail(gomock.Any(), email).Return(empData, nil)
				s.EXPECT().GetAbsenceReason(gomock.Any(), empData).Return(
					&domain.AbsenceReason{
						Status: "",
						Data: []domain.AbsenceReasonData{
							{
								Id:          28246,
								PersonId:    1234,
								CreatedDate: "2023-08-14",
								DateFrom:    "2023-08-12T00:00:00",
								DateTo:      "2023-08-12T23:59:59",
								ReasonId:    1,
							},
						},
					}, nil)
			},
			wantData: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович (🏠) Личные дела",
				Email:       "example@gmail.com",
				WorkPhone:   "+71234567890",
			},
			wantErr: nil,
		},
	}

	return &testTable
}

// todo СДЕЛАТЬ ТАК, ЧТО Я КАК БУДТО БЫ ПОЛУЧИЛ ОШИБКУ ИЗ СЕРВЕРА ВНЕШНЕГО НА ПЕРВОМ ИЛИ НА ВТОРОМ ЗАПРОСЕ.
func TestEmployeeService_GetReasonOfAbsence(t *testing.T) {
	testTable := *initTests()
	for _, tt := range testTable {
		t.Run(tt.testName, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockEmpRepo := mock_repository.NewMockEmployee(c)
			tt.mockBehavior(mockEmpRepo, tt.input.Email)
			reasons, err := domain.NewAbsenceOptions(domain.DefaultAbsenceOptionsFile)
			require.NoError(t, err)
			service := NewEmployeeService(mockEmpRepo, reasons)

			result, err := service.GetReasonOfAbsence(context.Background(), tt.input)

			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.wantData, result)
		})
	}
}
