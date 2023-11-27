package grpc

import (
	"context"
	"gRPCServer/internal/domain"
	dm "gRPCServer/internal/transport/grpc/sources/dataModification"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
)

func TestHandler_GetReasonOfAbsence(t *testing.T) {
	tests := []struct {
		testName       string
		input          *dm.ContactDetails
		respFromWorker domain.Future
		wantData       *dm.ContactDetails
		wantErr        error
	}{
		{
			testName: "OK",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "example@gmail.com",
				WorkPhone:   "+71234567890",
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
			wantErr:  domain.ErrInvalidEmail,
		},
		{
			testName: "incorrect email2",
			input: &dm.ContactDetails{
				DisplayName: "Иванов Семен Петрович",
				Email:       "23rijo09gw9",
				WorkPhone:   "+71234567890",
			},
			wantData: nil,
			wantErr:  domain.ErrInvalidEmail,
		},
	}

	workerResult := " (🏠) Личные дела"

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			handler := prepare()
			clientCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			input := tt.input

			go workerImitation(clientCtx, handler, workerResult)
			resp, err := handler.GetReasonOfAbsence(clientCtx, input)

			/*
				if !errors.Is(err, tt.wantErr) {
					t.Error("incorrect error", "expected: ", tt.wantErr, "got: ", err)
				}
				if err == nil {
					if !reflect.DeepEqual(resp, tt.wantData) {
						t.Error("incorrect data received", "expected: ", tt.wantData, "got: ", resp)
					}
				}
			*/
			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.wantData, resp)
		})
	}
}

func prepare() *Handler {
	channel := make(chan domain.AbsenceJob, 10)
	jq := domain.JobsQueue{AbsenceJQ: &channel}
	handler := NewHandler(grpc.NewServer(), jq)
	return handler
}

func workerImitation(ctx context.Context, handler *Handler, ans string) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-(*handler.JobsQueue.AbsenceJQ):
			if !ok {
				return
			}
			job.Input.DisplayName = job.Input.DisplayName + ans
			result := domain.Future{
				Error:  nil,
				Output: job.Input,
			}
			job.Result <- result
		}
	}
}
