package app

import (
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	mock_repository "gRPCServer/internal/repository/mocks"
	"gRPCServer/internal/server"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
	"gRPCServer/pkg/cache"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const logDir = "out"
const logLvl = "DEBUG"

var paths = domain.LoggerWritersPaths{
	GrpcTrafficFilePath: "grpcTrafficLog.txt",
	HttpTrafficFilePath: "httpTrafficLog.txt",
	ErrorWarnFilePath:   "errWarnLog.txt",
	DebugFilePath:       "debugLog.txt",
}

func Run(configDir string) error {
	cfg, err := config.ParseJsonConfig(configDir)
	if err != nil {
		return err
	}

	compositeLogger, err := domain.NewCompositeLogger(logDir, logLvl, paths)
	if err != nil {
		return err
	}

	channel := make(chan domain.AbsenceJob, cfg.AppServInfo.QueueSize)
	jq := domain.JobsQueue{AbsenceJQ: &channel}

	handler := transport.NewHandler(jq, compositeLogger)
	//EmpRepo := repository.NewEmployeeRepo(&cfg.ExtServInfo, compositeLogger)
	EmpRepo := &mock_repository.EmployeeRepoMock{}
	reasons := domain.NewAbsenceOptions()

	services := &service.Services{
		Employee: service.NewEmployeeService(EmpRepo, reasons,
			cache.NewMemoryCache(cfg.AppServInfo.TTLOfItemsInCache), compositeLogger),
	}
	s := server.NewServer(cfg, jq, handler, services, compositeLogger)

	go func() {
		if err := s.Run(); err != nil {
			log.Print(err)
		}
	}()
	log.Print("server is working")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit
	const timeout = 5 * time.Second
	log.Printf("shutting down the server, wait %d seconds", timeout/time.Second)
	s.GracefulStop(timeout)
	log.Print("server was stopped")
	return nil
}
