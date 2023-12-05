package tests

import (
	"context"
	"fmt"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	mock_repository "gRPCServer/internal/repository/mocks"
	"gRPCServer/internal/server"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

const configDir = "../configs"
const logDir = "out"
const logLvl = "DEBUG"

var paths = domain.LoggerWritersPaths{
	GrpcTrafficFilePath: "grpcTrafficLog.txt",
	HttpTrafficFilePath: "httpTrafficLog.txt",
	ErrorWarnFilePath:   "errWarnLog.txt",
	DebugFilePath:       "debugLog.txt",
}

func Test(t *testing.T) {
	cfg, err := config.ParseJsonConfig(configDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	compositeLogger, err := domain.NewCompositeLogger(logDir, logLvl, paths)
	if err != nil {
		return
	}

	channel := make(chan domain.AbsenceJob, cfg.AppServInfo.QueueSize)
	jq := domain.JobsQueue{AbsenceJQ: &channel}
	handler := transport.NewHandler(jq, compositeLogger)
	//EmpRepo := repository.NewEmployeeRepo(&cfg.ExtServInfo)
	//EmpRepo := mock_repository.NewMockEmployee()
	EmpRepo := &mock_repository.EmployeeRepoMock{}
	reasons := domain.NewAbsenceOptions()
	if err != nil {
		log.Fatal(err)
		return
	}
	services := &service.Services{
		Employee: service.NewEmployeeService(EmpRepo, reasons, compositeLogger),
	}
	s := server.NewServer(cfg, jq, handler, services, compositeLogger)

	go func() {
		s.Run()
	}()
	log.Print("server is working")

	go clientsRequestImitation()

	//time.Sleep(500 * time.Millisecond)
	const timeout = 5 * time.Second
	s.GracefulStop(timeout)
	log.Print("server was stopped")
}

const (
	address = "127.0.0.1:8080"
	n       = 100
)

// TODO сделать так, чтобы все запросы были со статусом OK.
// ну или хотя бы проверять что у них не было ошибки.
func clientsRequestImitation() {
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(p int) {
			defer wg.Done()
			time.Sleep(time.Nanosecond)
			request(p)
		}(i)
	}

	wg.Wait()
	fmt.Println("Done it")
}

func request(p int) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	c := dataModification.NewPersonalInfoClient(conn)

	cd := &dataModification.ContactDetails{
		DisplayName: "Иванов Семен Петрович",
		Email:       "example@gmail.com",
		MobilePhone: strconv.Itoa(p),
		WorkPhone:   strconv.Itoa(p),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := c.GetReasonOfAbsence(ctx, cd)
	if err != nil {
		log.Fatalf("some error: %v", err)
	}
	log.Print(
		"my data was: \n",
		cd.String(), "\n",
		"the data I get", "\n",
		resp.String(),
	)

}
