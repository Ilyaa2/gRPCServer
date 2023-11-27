package app

import (
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	"gRPCServer/internal/repository"
	"gRPCServer/internal/server"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const reasonsOptionsFileName = "reasons_options.txt"

// todo graceful stop (grpc) не согласован c пулом. Если grpc ожидает пока все соединения ослужатся в течение какого либо времени
func Run(configDir string) {
	cfg, err := config.ParseJsonConfig(configDir)
	if err != nil {
		log.Fatal(err)
		return
	}
	//todo нужно положить число в канале тоже в конфиг
	channel := make(chan domain.AbsenceJob, 10)
	jq := domain.JobsQueue{AbsenceJQ: &channel}
	handler := transport.NewHandler(grpc.NewServer(), jq)
	EmpRepo := repository.NewEmployeeRepo(&cfg.ExtServInfo)
	reasons, err := domain.NewAbsenceOptions(reasonsOptionsFileName)
	if err != nil {
		log.Fatal(err)
		return
	}
	services := &service.Services{
		Employee: service.NewEmployeeService(EmpRepo, reasons),
	}
	s := server.NewServer(cfg, jq, handler, services)

	go func() {
		s.Run()
	}()
	log.Print("server is working")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second
	s.GracefulStop(timeout)
	log.Print("server was stopped")
}
