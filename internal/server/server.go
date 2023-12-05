package server

import (
	"context"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
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

func (s *Server) Run() {
	address := s.Config.AppServInfo.ServerIp + ":" + s.Config.AppServInfo.ServerPort
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	}
	s.setWorkersPool()
	log.Printf("Starting gRPC listener on address: " + address)
	if err = s.handler.GrpcServ.Serve(lis); err != nil {
		s.GracefulStop(5 * time.Second)
		log.Fatalf("failed to serve: %v", err)
		return
	}
}

func (s *Server) GracefulStop(duration time.Duration) {
	s.handler.GrpcServ.GracefulStop()
	//to stop workers
	time.Sleep(duration)
	s.cancel()
	close(*s.JobsQueue.AbsenceJQ)
}

func (s *Server) Stop() {
	s.handler.GrpcServ.Stop()
	s.cancel()
	close(*s.JobsQueue.AbsenceJQ)
}

func (s *Server) setWorkersPool() {
	for i := 0; i < s.Config.AppServInfo.AmountOfWorkers; i++ {
		go func() {
			s.worker()
		}()
	}
}

func (s *Server) worker() {
	for {
		select {
		case <-s.globalCtx.Done():
			return
		case job, ok := <-(*s.JobsQueue.AbsenceJQ):
			if !ok {
				return
			}

			result, err := s.services.Employee.GetReasonOfAbsence(job.Context, job.Input)
			job.Result <- domain.Future{
				Error:  err,
				Output: result,
			}
		}
	}
}
