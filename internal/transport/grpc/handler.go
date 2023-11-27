package grpc

import (
	"context"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"google.golang.org/grpc"
)

type Handler struct {
	GrpcServ  *grpc.Server
	JobsQueue domain.JobsQueue
	dataModification.UnimplementedPersonalInfoServer
}

func NewHandler(g *grpc.Server, jq domain.JobsQueue) *Handler {
	h := &Handler{
		GrpcServ:  g,
		JobsQueue: jq,
	}
	dataModification.RegisterPersonalInfoServer(g, h)
	return h
}

// todo Контекст по сети может быть отменен - ctx.
// todo ФУНДАМЕНТАЛЬНО нужно провалидировать поля пользователя
func (h *Handler) GetReasonOfAbsence(ctx context.Context, data *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	if err := validateParameters(data); err != nil {
		return nil, domain.ErrInvalidEmail
	}
	result := make(chan domain.Future)
	job := domain.AbsenceJob{
		Input:  data,
		Result: result,
	}
	*h.JobsQueue.AbsenceJQ <- job
	future := <-result
	return future.Output, future.Error
}

func validateParameters(data *dataModification.ContactDetails) error {
	return validation.ValidateStruct(data, validation.Field(&data.Email, validation.Required, is.Email))
}
