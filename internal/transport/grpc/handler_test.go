package handler

import (
	"context"
	"gRPCServer/internal/domain"
	dm "gRPCServer/internal/transport/grpc/sources/dataModification"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

const logDir = "../../../out"
const logLvl = "DEBUG"

var paths = domain.LoggerWritersPaths{
	GrpcTrafficFilePath: "grpcTrafficLog.txt",
	HttpTrafficFilePath: "httpTrafficLog.txt",
	ErrorWarnFilePath:   "errWarnLog.txt",
	DebugFilePath:       "debugLog.txt",
}

const queueSize = 10

func prepare(t *testing.T) *Handler {
	channel := make(chan domain.AbsenceJob, queueSize)
	jq := domain.JobsQueue{AbsenceJQ: &channel}
	logs, err := domain.NewCompositeLogger(logDir, logLvl, paths)
	if err != nil {
		t.FailNow()
	}
	handler := NewHandler(jq, logs)
	return handler
}

func TestHandler_GetReasonOfAbsence_EmailValidation(t *testing.T) {
	tests := []struct {
		testName   string
		input      *dm.ContactDetails
		workerData domain.Future
		wantData   *dm.ContactDetails
		wantErr    error
	}{
		{
			testName: "OK",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "example@gmail.com",
				WorkPhone:   "+71234567890",
			},
			workerData: domain.Future{
				Error: nil,
				Output: &dm.ContactDetails{
					DisplayName: "Иванов Семен Петрович (🏠) Личные дела",
					Email:       "example@gmail.com",
					WorkPhone:   "+71234567890",
				},
			},
			wantData: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович (🏠) Личные дела",
				Email:       "example@gmail.com",
				WorkPhone:   "+71234567890",
			},
			wantErr: nil,
		},
		{
			testName: "incorrect email1",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "",
				WorkPhone:   "+71234567890",
			},
			wantData: nil,
			wantErr:  domain.ErrIncorrectEmail,
		},
		{
			testName: "incorrect email2",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "23rijo09gw9",
				WorkPhone:   "+71234567890",
			},
			wantData: nil,
			wantErr:  domain.ErrIncorrectEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := prepare(t)
			clientCtx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go workerImitation(clientCtx, handler, tt.workerData)
			resp, err := handler.GetReasonOfAbsence(clientCtx, tt.input)

			if err != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
			}
			require.Equal(t, tt.wantData, resp)
		})
	}
}

func TestHandler_GetReasonOfAbsence_ErrorFromWorker(t *testing.T) {
	tests := []struct {
		testName   string
		input      *dm.ContactDetails
		workerData domain.Future
		wantData   *dm.ContactDetails
		wantErr    error
	}{
		{
			testName: "error from worker 1",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "example@gmail.com",
				WorkPhone:   "+71234567890",
			},
			workerData: domain.Future{
				Error:  domain.ErrExternalServer,
				Output: nil,
			},
			wantData: nil,
			wantErr:  domain.ErrExternalServer,
		},
		{
			testName: "error from worker 2",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "example@gmail.com",
				WorkPhone:   "+71234567890",
			},
			workerData: domain.Future{
				Error:  domain.ErrNoInfoAvailable,
				Output: nil,
			},
			wantData: nil,
			wantErr:  domain.ErrNoInfoAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := prepare(t)
			clientCtx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go workerImitation(clientCtx, handler, tt.workerData)
			resp, err := handler.GetReasonOfAbsence(clientCtx, tt.input)

			if err != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
			}
			require.Equal(t, tt.wantData, resp)
		})
	}
}

func TestHandler_GetReasonOfAbsence_ErrorQueueIsFull(t *testing.T) {
	handler := prepare(t)
	for i := 0; i < queueSize; i++ {
		job := domain.AbsenceJob{
			Context: context.Background(),
			Input:   nil,
			Result:  make(chan domain.Future),
		}
		*handler.JobsQueue.AbsenceJQ <- job
	}
	test := struct {
		testName string
		input    *dm.ContactDetails
		wantErr  error
	}{
		testName: "queue is full error",
		input: &dm.ContactDetails{
			DisplayName: "Иванов Семен Петрович",
			Email:       "example@gmail.com",
			WorkPhone:   "+71234567890",
		},
		wantErr: status.Error(codes.Unavailable, ""),
	}

	t.Run(test.testName, func(t *testing.T) {
		resp, err := handler.GetReasonOfAbsence(context.Background(), test.input)
		require.ErrorContains(t, err, test.wantErr.Error())
		require.Nil(t, resp)
	})
}

func workerImitation(ctx context.Context, handler *Handler, future domain.Future) {
	select {
	case <-ctx.Done():
		return
	case job, ok := <-(*handler.JobsQueue.AbsenceJQ):
		if !ok {
			return
		}

		job.Result <- future
	}
}
