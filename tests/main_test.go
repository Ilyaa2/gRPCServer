package tests

import (
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	mock_repository "gRPCServer/internal/repository/mocks"
	"gRPCServer/internal/server"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func Test(t *testing.T) {
	configDir := "C:\\Users\\User\\GolandProjects\\gRPCServer\\configs"
	const reasonsOptionsFileName = "reasons_options.txt"
	cfg, err := config.ParseJsonConfig(configDir)
	if err != nil {
		log.Fatal(err)
		return
	}
	//todo нужно положить число в канале тоже в конфиг
	jq := make(chan domain.AbsenceJob, 10)
	handler := transport.NewHandler(grpc.NewServer(), &jq)
	//EmpRepo := repository.NewEmployeeRepo(&cfg.ExtServInfo)
	//EmpRepo := mock_repository.NewMockEmployee()
	EmpRepo := &mock_repository.EmployeeRepoMock{}
	reasons, err := domain.NewAbsenceOptions(reasonsOptionsFileName)
	if err != nil {
		log.Fatal(err)
		return
	}
	services := &service.Services{
		Employee: service.NewEmployeeService(EmpRepo, reasons),
	}
	s := server.NewServer(cfg, &jq, handler, services)

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
