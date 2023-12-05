package service

import (
	"context"
	"fmt"
	"gRPCServer/internal/domain"
	mock_repository "gRPCServer/internal/repository/mocks"
	dm "gRPCServer/internal/transport/grpc/sources/dataModification"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"testing"
)

const logDir = "../../out"
const logLvl = "DEBUG"

var paths = domain.LoggerWritersPaths{
	GrpcTrafficFilePath: "grpcTrafficLog.txt",
	HttpTrafficFilePath: "httpTrafficLog.txt",
	ErrorWarnFilePath:   "errWarnLog.txt",
	DebugFilePath:       "debugLog.txt",
}

func prepare(t *testing.T) domain.CompositeLogger {
	logs, err := domain.NewCompositeLogger(logDir, logLvl, paths)
	if err != nil {
		t.FailNow()
	}
	return logs
}

type mockBehavior func(*mock_repository.MockEmployee, string)

type Tests struct {
	testName     string
	input        *dm.ContactDetails
	mockBehavior mockBehavior
	wantData     *dm.ContactDetails
	wantErr      error
}

func TestEmployeeService_GetReasonOfAbsence_SuccessfulWithEmoji(t *testing.T) {
	test := Tests{
		testName: "Successful with emoji",
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
					Status: "OK",
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
			DisplayName: "Иванов Семен Петрович - (🏠) Личные дела",
			Email:       "example@gmail.com",
			WorkPhone:   "+71234567890",
		},
		wantErr: nil,
	}

	t.Run(test.testName, func(t *testing.T) {
		c := gomock.NewController(t)
		defer c.Finish()

		mockEmpRepo := mock_repository.NewMockEmployee(c)
		test.mockBehavior(mockEmpRepo, test.input.Email)
		logs := prepare(t)
		reasons := domain.NewAbsenceOptions()
		service := NewEmployeeService(mockEmpRepo, reasons, logs)

		result, err := service.GetReasonOfAbsence(context.Background(), test.input)

		require.Equal(t, test.wantErr, err)
		require.Equal(t, test.wantData, result)
	})
}

func TestEmployeeService_GetReasonOfAbsence_SuccessfulWithoutEmoji(t *testing.T) {
	test := Tests{
		testName: "Successful without emoji",
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
					Status: "OK",
					Data: []domain.AbsenceReasonData{
						{
							Id:          28246,
							PersonId:    1234,
							CreatedDate: "2023-08-14",
							DateFrom:    "2023-08-12T00:00:00",
							DateTo:      "2023-08-12T23:59:59",
							ReasonId:    8,
						},
					},
				}, nil)
		},
		wantData: &dm.ContactDetails{
			DisplayName: "Иванов Семен Петрович - Дежурство",
			Email:       "example@gmail.com",
			WorkPhone:   "+71234567890",
		},
		wantErr: nil,
	}

	t.Run(test.testName, func(t *testing.T) {
		c := gomock.NewController(t)
		defer c.Finish()

		mockEmpRepo := mock_repository.NewMockEmployee(c)
		test.mockBehavior(mockEmpRepo, test.input.Email)
		logs := prepare(t)
		reasons := domain.NewAbsenceOptions()
		service := NewEmployeeService(mockEmpRepo, reasons, logs)

		result, err := service.GetReasonOfAbsence(context.Background(), test.input)

		require.Equal(t, test.wantErr, err)
		require.Equal(t, test.wantData, result)
	})
}

func TestEmployeeService_GetReasonOfAbsence_RepoError1(t *testing.T) {
	test := Tests{
		testName: "Error in repo function: GetByEmail",
		input: &dm.ContactDetails{
			DisplayName: "Иванов Семен Петрович",
			Email:       "example@gmail.com",
			WorkPhone:   "+71234567890",
		},
		mockBehavior: func(s *mock_repository.MockEmployee, email string) {
			s.EXPECT().GetByEmail(gomock.Any(), email).Return(nil,
				fmt.Errorf("%w: status code of responce = %v", domain.ErrExternalServer, http.StatusNotFound))
		},
		wantData: nil,
		wantErr:  domain.ErrExternalServer,
	}

	t.Run(test.testName, func(t *testing.T) {
		c := gomock.NewController(t)
		defer c.Finish()

		mockEmpRepo := mock_repository.NewMockEmployee(c)
		test.mockBehavior(mockEmpRepo, test.input.Email)
		logs := prepare(t)
		reasons := domain.NewAbsenceOptions()
		service := NewEmployeeService(mockEmpRepo, reasons, logs)

		result, err := service.GetReasonOfAbsence(context.Background(), test.input)

		require.ErrorIs(t, err, test.wantErr)
		require.Equal(t, test.wantData, result)
	})
}

func TestEmployeeService_GetReasonOfAbsence_RepoError2(t *testing.T) {
	test := Tests{
		testName: "Error in repo function: GetAbsenceReason",
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
					Status: "NOT FOUND",
					Data:   nil,
				}, nil)
		},
		wantData: nil,
		wantErr:  domain.ErrNoInfoAvailable,
	}

	t.Run(test.testName, func(t *testing.T) {
		c := gomock.NewController(t)
		defer c.Finish()

		mockEmpRepo := mock_repository.NewMockEmployee(c)
		test.mockBehavior(mockEmpRepo, test.input.Email)
		logs := prepare(t)
		reasons := domain.NewAbsenceOptions()
		service := NewEmployeeService(mockEmpRepo, reasons, logs)

		result, err := service.GetReasonOfAbsence(context.Background(), test.input)

		require.ErrorIs(t, err, test.wantErr)
		require.Equal(t, test.wantData, result)
	})
}

func TestEmployeeService_GetReasonOfAbsence_NoReasonIdInDict(t *testing.T) {
	test := Tests{
		testName: "Absence options doesn't have reason id given",
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
					Status: "OK",
					Data: []domain.AbsenceReasonData{
						{
							Id:          28246,
							PersonId:    1234,
							CreatedDate: "2023-08-14",
							DateFrom:    "2023-08-12T00:00:00",
							DateTo:      "2023-08-12T23:59:59",
							ReasonId:    100,
						},
					},
				}, nil)
		},
		wantData: &dm.ContactDetails{
			DisplayName: "Иванов Семен Петрович",
			Email:       "example@gmail.com",
			WorkPhone:   "+71234567890",
		},
		wantErr: nil,
	}

	t.Run(test.testName, func(t *testing.T) {
		c := gomock.NewController(t)
		defer c.Finish()

		mockEmpRepo := mock_repository.NewMockEmployee(c)
		test.mockBehavior(mockEmpRepo, test.input.Email)
		logs := prepare(t)
		reasons := domain.NewAbsenceOptions()
		service := NewEmployeeService(mockEmpRepo, reasons, logs)

		result, err := service.GetReasonOfAbsence(context.Background(), test.input)

		require.Equal(t, test.wantErr, err)
		require.Equal(t, test.wantData, result)
	})
}
