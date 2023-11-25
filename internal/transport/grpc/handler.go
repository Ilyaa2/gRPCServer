package grpc

import (
	"context"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	"google.golang.org/grpc"
)

type Handler struct {
	GrpcServ  *grpc.Server
	JobsQueue *chan domain.AbsenceJob
	dataModification.UnimplementedPersonalInfoServer
}

func NewHandler(g *grpc.Server, jQueue *chan domain.AbsenceJob) *Handler {
	h := &Handler{
		GrpcServ:  g,
		JobsQueue: jQueue,
	}
	dataModification.RegisterPersonalInfoServer(g, h)
	return h
}

// todo Контекст по сети может быть отменен - ctx.
// todo ФУНДАМЕНТАЛЬНО нужно провалидировать поля пользователя
func (s *Handler) GetReasonOfAbsence(ctx context.Context, data *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	result := make(chan string)
	job := domain.AbsenceJob{
		Data:   data,
		Result: result,
	}
	*s.JobsQueue <- job

	data.DisplayName = data.Email + <-result
	return data, nil
}
