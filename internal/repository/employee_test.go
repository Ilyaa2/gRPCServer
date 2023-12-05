package repository

import (
	"context"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const logDir = "../../out"
const logLvl = "DEBUG"

var paths = domain.LoggerWritersPaths{
	GrpcTrafficFilePath: "grpcTrafficLog.txt",
	HttpTrafficFilePath: "httpTrafficLog.txt",
	ErrorWarnFilePath:   "errWarnLog.txt",
	DebugFilePath:       "debugLog.txt",
}

type emailTests struct {
	employeeServer *httptest.Server
	wantedData     *domain.EmployeeData
	input          string
	wantedErr      error
}

func prepare(t *testing.T, employeeUrlPath, absenceUrlPath string) Employee {
	cfg := config.Config{
		ExtServInfo: config.ExternalServerInfo{
			EmployeeUrlPath: employeeUrlPath,
			AbsenceUrlPath:  absenceUrlPath,
			Login:           "login",
			Password:        "password",
			RequestTimeout:  1000,
		},
	}
	logs, err := domain.NewCompositeLogger(logDir, logLvl, paths)
	if err != nil {
		t.FailNow()
	}
	return NewEmployeeRepo(&cfg.ExtServInfo, logs)
}

func TestEmployeeRepo_GetByEmail_Successful(t *testing.T) {
	test := emailTests{
		employeeServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response :=
				`{
					"status": "OK",
					"data": [
						{
							"id": 1234,
							"displayName": "Иванов Семен Петрович",
							"email": "petrovich@mail.ru",
							"workPhone": "+71234567890"
						}
					  ]
					}`
			_, _ = w.Write([]byte(response))
		})),
		wantedData: &domain.EmployeeData{
			Status: "OK",
			Data: []domain.EmployeeInnerData{
				{
					Id:          1234,
					DisplayName: "Иванов Семен Петрович",
					Email:       "petrovich@mail.ru",
					WorkPhone:   "+71234567890",
				},
			},
		},
		wantedErr: nil,
		input:     "petrovich@mail.ru",
	}

	t.Run("SUCCESSFUL", func(t *testing.T) {
		defer test.employeeServer.Close()

		repo := prepare(t, test.employeeServer.URL, "")
		empData, err := repo.GetByEmail(context.Background(), test.input)
		require.Equal(t, test.wantedErr, err)
		require.Equal(t, test.wantedData, empData)
	})
}

func TestEmployeeRepo_GetByEmail_DeadlineExceeded(t *testing.T) {
	test := emailTests{
		employeeServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(4 * time.Second)
			w.WriteHeader(http.StatusOK)
			response :=
				`{
					"status": "OK",
					"data": [
						{
							"id": 1234,
							"displayName": "Иванов Семен Петрович",
							"email": "petrovich@mail.ru",
							"workPhone": "+71234567890"
						}
					  ]
					}`
			_, _ = w.Write([]byte(response))
		})),
		wantedData: nil,
		wantedErr:  domain.ErrExternalServer,
		input:      "petrovich@mail.ru",
	}

	t.Run("Deadline was exceeded", func(t *testing.T) {
		defer test.employeeServer.Close()

		repo := prepare(t, test.employeeServer.URL, "")
		empData, err := repo.GetByEmail(context.Background(), test.input)
		require.ErrorIs(t, err, test.wantedErr)
		require.Equal(t, test.wantedData, empData)
	})
}

func TestEmployeeRepo_GetByEmail_ExtServerDoesNotHaveExactEmail(t *testing.T) {
	test := emailTests{
		employeeServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			response :=
				`{
					"status": "NOT FOUND"
				}`
			_, _ = w.Write([]byte(response))
		})),
		wantedData: nil,
		wantedErr:  domain.ErrExternalServer,
		input:      "petrovich@mail.ru",
	}

	t.Run("External server doesn't have this email", func(t *testing.T) {
		defer test.employeeServer.Close()

		repo := prepare(t, test.employeeServer.URL, "")
		empData, err := repo.GetByEmail(context.Background(), test.input)
		require.ErrorIs(t, err, test.wantedErr)
		require.Equal(t, test.wantedData, empData)
	})
}

func TestEmployeeRepo_GetByEmail_WrongCredentials(t *testing.T) {
	test := emailTests{
		employeeServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			response :=
				`{
					"status": "WRONG CREDENTIALS"
				}`
			_, _ = w.Write([]byte(response))
		})),
		wantedData: nil,
		wantedErr:  domain.ErrExternalServer,
		input:      "petrovich@mail.ru",
	}

	t.Run("Email or password is wrong for External Server", func(t *testing.T) {
		defer test.employeeServer.Close()

		repo := prepare(t, test.employeeServer.URL, "")
		empData, err := repo.GetByEmail(context.Background(), test.input)
		require.ErrorIs(t, err, test.wantedErr)
		require.Equal(t, test.wantedData, empData)
	})
}

type absenceTest struct {
	absenceServer *httptest.Server
	wantedData    *domain.AbsenceReason
	input         *domain.EmployeeData
	wantedErr     error
}

func TestEmployeeRepo_GetAbsenceReason_Successful(t *testing.T) {
	test := absenceTest{
		absenceServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response :=
				`{
					"status": "OK",
					"data": [
						{
							"id": 28246,
							"personId": 1234,
							"createdDate": "2023-08-14",
							"dateFrom": "2023-08-12T00:00:00",
							"dateTo": "2023-08-12T23:59:59",
							"reasonId": 1
						}
                      ]
					}`
			w.Write([]byte(response))
		})),
		input: &domain.EmployeeData{
			Status: "OK",
			Data: []domain.EmployeeInnerData{
				{
					Id:          1234,
					DisplayName: "Иванов Семен Петрович",
					Email:       "petrovich@mail.ru",
					WorkPhone:   "+71234567890",
				},
			},
		},
		wantedData: &domain.AbsenceReason{
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
		},
		wantedErr: nil,
	}

	t.Run("SUCCESSFUL", func(t *testing.T) {
		defer test.absenceServer.Close()

		repo := prepare(t, "", test.absenceServer.URL)
		empData, err := repo.GetAbsenceReason(context.Background(), test.input)
		require.ErrorIs(t, err, test.wantedErr)
		require.Equal(t, test.wantedData, empData)
	})
}

func TestEmployeeRepo_GetAbsenceReason_EmptyDataFromExternalServer(t *testing.T) {
	test := absenceTest{
		absenceServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response :=
				`{
					"status": "OK"
				}`
			w.Write([]byte(response))
		})),
		input: &domain.EmployeeData{
			Status: "OK",
			Data: []domain.EmployeeInnerData{
				{
					Id:          1234,
					DisplayName: "Иванов Семен Петрович",
					Email:       "petrovich@mail.ru",
					WorkPhone:   "+71234567890",
				},
			},
		},
		wantedData: &domain.AbsenceReason{
			Status: "OK",
			Data:   nil,
		},
		wantedErr: nil,
	}

	t.Run("NO data from external server", func(t *testing.T) {
		defer test.absenceServer.Close()

		repo := prepare(t, "", test.absenceServer.URL)
		empData, err := repo.GetAbsenceReason(context.Background(), test.input)
		require.ErrorIs(t, err, test.wantedErr)
		require.Equal(t, test.wantedData, empData)
	})
}

func TestEmployeeRepo_GetAbsenceReason_DeadlineExceeded(t *testing.T) {
	test := absenceTest{
		absenceServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Second * 4)
			w.WriteHeader(http.StatusOK)
			response :=
				`{
					"status": "OK",
					"data": [
						{
							"id": 28246,
							"personId": 1234,
							"createdDate": "2023-08-14",
							"dateFrom": "2023-08-12T00:00:00",
							"dateTo": "2023-08-12T23:59:59",
							"reasonId": 1
						}
                      ]
					}`
			w.Write([]byte(response))
		})),
		input: &domain.EmployeeData{
			Status: "OK",
			Data: []domain.EmployeeInnerData{
				{
					Id:          1234,
					DisplayName: "Иванов Семен Петрович",
					Email:       "petrovich@mail.ru",
					WorkPhone:   "+71234567890",
				},
			},
		},
		wantedData: nil,
		wantedErr:  domain.ErrExternalServer,
	}

	t.Run("Deadline was exceeded", func(t *testing.T) {
		defer test.absenceServer.Close()

		repo := prepare(t, "", test.absenceServer.URL)
		empData, err := repo.GetAbsenceReason(context.Background(), test.input)
		require.ErrorIs(t, err, test.wantedErr)
		require.Equal(t, test.wantedData, empData)
	})
}

func TestEmployeeRepo_GetAbsenceReason_WrongCredentials(t *testing.T) {
	test := absenceTest{
		absenceServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			response :=
				`{
					"status": "WRONG CREDENTIALS"
				}`
			w.Write([]byte(response))
		})),
		input: &domain.EmployeeData{
			Status: "OK",
			Data: []domain.EmployeeInnerData{
				{
					Id:          1234,
					DisplayName: "Иванов Семен Петрович",
					Email:       "petrovich@mail.ru",
					WorkPhone:   "+71234567890",
				},
			},
		},
		wantedData: nil,
		wantedErr:  domain.ErrExternalServer,
	}

	t.Run("Email or password is wrong for External Server", func(t *testing.T) {
		defer test.absenceServer.Close()

		repo := prepare(t, "", test.absenceServer.URL)
		empData, err := repo.GetAbsenceReason(context.Background(), test.input)
		require.ErrorIs(t, err, test.wantedErr)
		require.Equal(t, test.wantedData, empData)
	})
}
