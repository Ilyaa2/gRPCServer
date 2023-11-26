package server

import (
	"context"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
	"log"
	"math/rand"
	"net"
	"time"
)

type Server struct {
	Config *config.Config
	//todo это надо куда-то спрятать
	JobsQueue *chan domain.AbsenceJob
	ctx       context.Context
	cancel    context.CancelFunc
	handler   *transport.Handler
	services  *service.Services
}

// todo передавать handler сверху в app.Run().
func NewServer(cfg *config.Config, jq *chan domain.AbsenceJob, handler *transport.Handler, services *service.Services) *Server {
	//todo с контекстом потом нужно разобраться.
	ctx, cancel := context.WithCancel(context.Background())

	s := &Server{
		handler:   handler,
		JobsQueue: jq,
		Config:    cfg,
		services:  services,
		ctx:       ctx,
		cancel:    cancel,
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
}

// todo если случилась ошибка на запросе клиента, нужно сделать так, чтоб горутина не умирала.
func (s *Server) setWorkersPool() {
	for i := 0; i < s.Config.AppServInfo.AmountOfWorkers; i++ {
		go func() {
			s.worker()
		}()
	}
}

// todo нужно сделать так, что
// todo хотя с другой стороны сервер может просто закрыть канал.
// todo причем как сервер может отменить контекст - тогда действительно воркеры не работают, так и
// todo getReason может отменять контекст и тогда у меня воркер просто возвращается опять в пул.
func (s *Server) worker() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case job, ok := <-(*s.JobsQueue):
			if !ok {
				return
			}
			//todo правильный ли контекст даю? если getReasonAbsence отменит контекст, то воркер умрет
			result, err := s.services.Employee.GetReasonOfAbsence(context.Background(), job.Input)
			job.Result <- domain.Future{
				Error:  err,
				Output: result,
			}
		}
	}
}

func imitateProcess() string {
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 3)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
