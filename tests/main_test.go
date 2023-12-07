package tests

import (
	"context"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	mock_repository "gRPCServer/internal/repository/mocks"
	"gRPCServer/internal/server"
	"gRPCServer/internal/service"
	transport "gRPCServer/internal/transport/grpc"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	"gRPCServer/pkg/cache"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"strconv"
	"sync"
	"testing"
	"time"
)

const configDir = "../configs"
const logDir = "../out"
const logLvl = "DEBUG"

var paths = domain.LoggerWritersPaths{
	GrpcTrafficFilePath: "grpcTrafficLog.txt",
	HttpTrafficFilePath: "httpTrafficLog.txt",
	ErrorWarnFilePath:   "errWarnLog.txt",
	DebugFilePath:       "debugLog.txt",
}

func TestWholeApp(t *testing.T) {
	cfg, err := config.ParseJsonConfig(configDir)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	compositeLogger, err := domain.NewCompositeLogger(logDir, logLvl, paths)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	channel := make(chan domain.AbsenceJob, cfg.AppServInfo.QueueSize)
	jq := domain.JobsQueue{AbsenceJQ: &channel}
	handler := transport.NewHandler(jq, compositeLogger)
	EmpRepo := &mock_repository.EmployeeRepoMock{}
	reasons := domain.NewAbsenceOptions()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	services := &service.Services{
		Employee: service.NewEmployeeService(EmpRepo, reasons,
			cache.NewMemoryCache(cfg.AppServInfo.TTLOfItemsInCache), compositeLogger),
	}
	s := server.NewServer(cfg, jq, handler, services, compositeLogger)

	go func() {
		err := s.Run()
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
	}()
	t.Log("server is working")

	t.Run("requests to server", func(t *testing.T) {
		go clientsRequestImitation(t)

		time.Sleep(500 * time.Millisecond)
		s.Stop()
		t.Log("server was stopped")
	})
}

const (
	address = "127.0.0.1:8080"
	n       = 100
)

func clientsRequestImitation(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(p int) {
			defer wg.Done()
			time.Sleep(time.Nanosecond)
			request(p, t)
		}(i)
	}

	wg.Wait()
	t.Log("Done requests")
}

func request(p int, t *testing.T) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	c := dataModification.NewPersonalInfoClient(conn)

	input := &dataModification.ContactDetails{
		DisplayName: "Иванов Семен Петрович",
		Email:       "example" + strconv.Itoa(p) + "@gmail.com",
		WorkPhone:   "",
	}
	wantedData := &dataModification.ContactDetails{
		DisplayName: "Иванов Семен Петрович - (🏠) Личные дела",
		Email:       "example" + strconv.Itoa(p) + "@gmail.com",
		WorkPhone:   "",
	}
	wantedErr := codes.Unavailable
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := c.GetReasonOfAbsence(ctx, input)

	if err != nil {
		require.ErrorContains(t, err, wantedErr.String())
	} else {
		if wantedData.Email != resp.Email || wantedData.DisplayName != resp.DisplayName {
			t.Logf("the data wanted %v\n the data I get %v\n", wantedData, resp)
			t.FailNow()
		}
	}
}
