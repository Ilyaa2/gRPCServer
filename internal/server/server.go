package server

import (
	"context"
	"fmt"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
	"gRPCServer/pkg/util"
	"log"
	"net"
	"time"
)

type Server struct {
	Config          *config.Config
	JobsQueue       domain.JobsQueue
	globalCtx       context.Context
	cancel          context.CancelFunc
	handler         *transport.Handler
	services        *service.Services
	compositeLogger domain.CompositeLogger
}

func NewServer(cfg *config.Config, jq domain.JobsQueue, handler *transport.Handler,
	services *service.Services, logger domain.CompositeLogger) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		handler:         handler,
		JobsQueue:       jq,
		Config:          cfg,
		services:        services,
		globalCtx:       ctx,
		cancel:          cancel,
		compositeLogger: logger,
	}
	return s
}

// Run launch worker pool and grpc server. If an error appears then server will stop it gracefully.
func (s *Server) Run() error {
	address := s.Config.AppServInfo.ServerIp + ":" + s.Config.AppServInfo.ServerPort
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = fmt.Errorf("failed to listen: %v", err)
		s.compositeLogger.ApplicationLogger.Error(
			"",
			map[string]interface{}{
				"package":  "server",
				"function": "Run",
				"err":      err,
			},
		)
		return err
	}
	s.setWorkersPool()
	log.Printf("Starting gRPC listener on address: " + address)
	if err = s.handler.GrpcServ.Serve(lis); err != nil {
		s.GracefulStop(5 * time.Second)
		err = fmt.Errorf("failed to serve: %v", err)
		s.compositeLogger.ApplicationLogger.Error(
			"",
			map[string]interface{}{
				"package":  "server",
				"function": "Run",
				"err":      err,
			},
		)
		return err
	}
	return nil
}

func (s *Server) GracefulStop(duration time.Duration) {
	s.handler.GrpcServ.GracefulStop()
	//to stop workers
	time.Sleep(duration)
	s.cancel()
	close(*s.JobsQueue.AbsenceJQ)
	s.compositeLogger.ApplicationLogger.Warn("shutting down the server gracefully", nil)
}

func (s *Server) Stop() {
	s.handler.GrpcServ.Stop()
	s.cancel()
	close(*s.JobsQueue.AbsenceJQ)
	s.compositeLogger.ApplicationLogger.Warn("shutting down the server", nil)
}

func (s *Server) setWorkersPool() {
	for i := 0; i < s.Config.AppServInfo.AmountOfWorkers; i++ {
		go func() {
			s.worker()
		}()
	}
}

// worker takes the task in JobsQueue. Task was created by handler which accepts all requests via grpc.
// After processing worker put the request in the Future channel.
// Handler was blocked on this channel all that processing time.
func (s *Server) worker() {
	for {
		select {
		case <-s.globalCtx.Done():
			return
		case job, ok := <-(*s.JobsQueue.AbsenceJQ):
			if !ok {
				return
			}
			s.compositeLogger.RequestResponseLogger.Debug(
				"worker got the job",
				map[string]interface{}{
					"req-id":   util.GetReqIDFromContext(job.Context),
					"package":  "server",
					"function": "worker",
					"job":      job.Input,
				},
			)
			result, err := s.services.Employee.GetReasonOfAbsence(job.Context, job.Input)
			s.compositeLogger.RequestResponseLogger.Debug(
				"worker have done the job",
				map[string]interface{}{
					"req-id":   util.GetReqIDFromContext(job.Context),
					"package":  "server",
					"function": "worker",
					"result":   result,
					"err":      err,
				},
			)
			job.Result <- domain.Future{
				Error:  err,
				Output: result,
			}
		}
	}
}
