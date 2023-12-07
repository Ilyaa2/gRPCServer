package handler

import (
	"context"
	"errors"
	"fmt"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Handler struct {
	GrpcServ  *grpc.Server
	JobsQueue domain.JobsQueue
	dataModification.UnimplementedPersonalInfoServer
	compositeLogger domain.CompositeLogger
}

func (h *Handler) unaryLogInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	reqID := uuid.New().String()
	childCtx := context.WithValue(ctx, "req-id", reqID)

	h.compositeLogger.RequestResponseLogger.Debug(
		"request",
		map[string]interface{}{
			"req-id":  reqID,
			"method":  info.FullMethod,
			"request": req,
		},
	)

	resp, err := handler(childCtx, req)

	logMap := map[string]interface{}{
		"req-id":     reqID,
		"rpc-method": info.FullMethod,
		"response":   resp,
	}
	if err != nil {
		h.compositeLogger.RequestResponseLogger.Debug(
			"response",
			logMap,
		)
	} else {
		logMap["err"] = err
		h.compositeLogger.RequestResponseLogger.Error(
			"response",
			logMap,
		)
	}

	return resp, err
}

func NewHandler(jq domain.JobsQueue, logger domain.CompositeLogger) *Handler {
	h := &Handler{
		JobsQueue:       jq,
		compositeLogger: logger,
	}
	h.GrpcServ = grpc.NewServer(grpc.UnaryInterceptor(h.unaryLogInterceptor))
	dataModification.RegisterPersonalInfoServer(h.GrpcServ, h)
	return h
}

func errWithGrpcCodes(err error) error {
	switch {
	case err == nil:
		return status.Error(codes.OK, "")
	case errors.Is(err, domain.ErrIncorrectEmail):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNoInfoAvailable), errors.Is(err, domain.ErrNilData):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrViolatedJsonContract), errors.Is(err, domain.ErrExternalServer):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return status.Error(codes.FailedPrecondition, err.Error())
	}
}

func (h *Handler) GetReasonOfAbsence(ctx context.Context, data *dataModification.ContactDetails) (*dataModification.ContactDetails, error) {
	reqID, ok := ctx.Value("req-id").(string)
	if !ok {
		reqID = "none"
	}
	if err := validateParameters(data); err != nil {
		err = fmt.Errorf("%w. Details: %v", domain.ErrIncorrectEmail, err)
		h.compositeLogger.ApplicationLogger.Warn(
			"incorrect email from user",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "handler",
				"function": "GetReasonOfAbsence",
				"err":      err,
			},
		)
		return nil, errWithGrpcCodes(err)
	}

	result := make(chan domain.Future)
	job := domain.AbsenceJob{
		Context: ctx,
		Input:   data,
		Result:  result,
	}

	h.compositeLogger.ApplicationLogger.Debug("Characteristics of Queue of jobs",
		map[string]interface{}{
			"Size of the queue: ": len(*h.JobsQueue.AbsenceJQ),
			"Capacity:":           cap(*h.JobsQueue.AbsenceJQ),
		})

	select {
	case *h.JobsQueue.AbsenceJQ <- job:
		h.compositeLogger.ApplicationLogger.Debug(
			"handler have put the job in the queue",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "handler",
				"function": "GetReasonOfAbsence",
				"job":      job.Input,
			})

		future := <-result

		h.compositeLogger.ApplicationLogger.Debug(
			"handler got the result from the queue",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "handler",
				"function": "GetReasonOfAbsence",
				"future":   future.Output,
				"err":      future.Error,
			})
		return future.Output, errWithGrpcCodes(future.Error)
	default:
		stat := status.New(codes.Unavailable, "We are experiencing a large number of requests. Try later")
		detailedStatus, _ := stat.WithDetails(
			&errdetails.RetryInfo{RetryDelay: &durationpb.Duration{
				Seconds: 60,
			},
			},
		)
		h.compositeLogger.ApplicationLogger.Warn(
			"Don't have space in the queue to process a request", map[string]interface{}{"req-id": reqID},
		)
		return nil, detailedStatus.Err()
	}
}

func validateParameters(data *dataModification.ContactDetails) error {
	return validation.ValidateStruct(data, validation.Field(&data.Email, validation.Required, is.Email))
}
