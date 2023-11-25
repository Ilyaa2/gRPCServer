package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	transport "gRPCServer/internal/transport/grpc"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	//todo не знаю куда бы запрятать JobsQueue
	//services interface{} - можно как дать ему JobsQueue, так и чтоб worker'ы давали сверху необходимые поля
	Config *config.Config
	//todo это надо куда-то спрятать
	//во вторых хотелось, чтобы она не была привязана только к одной работе.
	JobsQueue chan domain.AbsenceJob
	ctx       context.Context
	cancel    context.CancelFunc
	handler   *transport.Handler
}

// todo передавать handler сверху в app.Run().
func NewServer(cfg *config.Config) *Server {
	//todo с контекстом потом нужно разобраться.
	//ctx, cancel := context.WithCancel(context.Background())
	jq := make(chan domain.AbsenceJob, 10)
	s := &Server{
		handler: transport.NewHandler(grpc.NewServer(), &jq),
		//todo нужно положить число в канале тоже в конфиг
		JobsQueue: jq,
		Config:    cfg,
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
	if err := s.handler.GrpcServ.Serve(lis); err != nil {
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

// todo нужно сделать так, что если getReasonAbsence отменит контекст, то воркер не умирал, а шел обратно в пул.
// todo хотя с другой стороны сервер может просто закрыть канал.
// todo причем как сервер может отменить контекст - тогда действительно воркеры не работают, так и
// todo getReason может отменять контекст и тогда у меня воркер просто возвращается опять в пул.
func (s *Server) worker() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case job, ok := <-s.JobsQueue:
			if !ok {
				return
			}
			//todo нужен ли сон???
			time.Sleep(time.Millisecond)
			resultName := job.Data.Email + imitateProcess()
			job.Result <- resultName
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

//ниже это все логика для service.
//==============================================================

type employeeData struct {
	Status string `json:"status"`
	Data   []struct {
		Id          int    `json:"id"`
		DisplayName string `json:"displayName"`
		Email       string `json:"email"`
		WorkPhone   string `json:"workPhone"`
	} `json:"data"`
}

type absenceReason struct {
	Status string `json:"status"`
	Data   []struct {
		Id          int    `json:"id"`
		PersonId    int    `json:"personId"`
		CreatedDate string `json:"createdDate"`
		DateFrom    string `json:"dateFrom"`
		DateTo      string `json:"dateTo"`
		ReasonId    int    `json:"reasonId"`
	} `json:"data"`
}

// todo нужно создать типы ошибок
// todo ФУНДАМЕНТАЛЬНО НЕПРАВИЛЬНО должен возвращать измененный dm.ContactDetails. В ФИО дописать через словарь причину отсутствия. Нужно ли блокироваться при поиске в словаре?
// todo Скорее всего нужно будет реализовать свою Unmodifiable/Immutable map - просто свою структуру.
func (s *Server) reasonOfAbsence(ctx context.Context, details *dataModification.ContactDetails) (int, error) {
	empData, err := s.employeeDataRequest(ctx, details.Email)
	if err == nil {
		return 0, err
	}
	// добавил только один id
	absReason, err := s.reasonOfAbsenceRequest(ctx, empData)
	if err != nil {
		return 0, err
	}
	return absReason.Data[0].ReasonId, nil
}

// todo так же чекай текущий контекст
// todo установи таймаут на соединение с сервером. Если таймаут истек - верни ошибку.
// todo хардкод в url -> сделать так чтоб в конфиге это задавалось
// todo тут или в doRequest нужно закрывать resp или resp.Body (метод close)
func (s *Server) employeeDataRequest(ctx context.Context, email string) (*employeeData, error) {
	requestInfo := func() (string, interface{}) {
		type reqBody struct {
			Email    string    `json:"email"`
			DateFrom time.Time `json:"dateFrom"`
			DateTo   time.Time `json:"dateTo"`
		}

		t := time.Now()
		rb := reqBody{
			Email:    email,
			DateFrom: t,
			DateTo:   t.Add(time.Hour*24 - time.Millisecond),
		}
		url := "https://" + s.Config.ExtServInfo.Ip + ":" + s.Config.ExtServInfo.Port + "/Portal/springApi/api/employees"
		return url, rb
	}
	resp, err := s.doRequest(ctx, requestInfo)
	if err != nil {
		return nil, err
	}
	respData := &employeeData{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
		err = errors.New("The json contract between application and external server is violated: " + err.Error())
		return nil, err
	}
	return respData, nil
}

// todo смотреть за контекстом, его могут отменить
// todo при запросе установить дедлайн при котором сервер должен ответить. Дедлайн брать из конфигурации
func (s *Server) doRequest(ctx context.Context, requestInfo func() (string, interface{})) (*http.Response, error) {
	url, requestBody := requestInfo()
	jsonReq, err := json.Marshal(requestBody)
	if err != nil {
		errors.New("Error during the conversion into JSON: " + err.Error())
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonReq))
	if err != nil {
		err = errors.New("Error during the creation of request: " + err.Error())
		return nil, err
	}

	req.SetBasicAuth(s.Config.ExtServInfo.Login, s.Config.ExtServInfo.Password)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errors.New("Error during request execution. External server might not be available: " + err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errors.New("Error on the external server side. Status code = " + strconv.Itoa(resp.StatusCode) + ", Status: " + resp.Status)
		return nil, err
	}

	return resp, nil
}

// todo возможно, если пользователю нужно возвращать id, то нужно создать новую сущность, а не только contactDetails
// todo хардкод в url -> сделать так чтоб в конфиге это задавалось
// todo так же чекай текущий контекст
// todo []int{empData.Data[0].Id}. Разобраться при первом запросе можно ли послать массив email или нужно много делать запросов по всем email. Задать этот вопрос
func (s *Server) reasonOfAbsenceRequest(ctx context.Context, empData *employeeData) (*absenceReason, error) {
	requestInfo := func() (string, interface{}) {
		type reqBody struct {
			PersonIds []int     `json:"personIds"`
			DateFrom  time.Time `json:"dateFrom"`
			DateTo    time.Time `json:"dateTo"`
		}

		t := time.Now()
		rb := reqBody{
			//todo добавил только один id
			PersonIds: []int{empData.Data[0].Id},
			DateFrom:  t,
			DateTo:    t.Add(time.Hour*24 - time.Millisecond),
		}
		url := "https://" + s.Config.ExtServInfo.Ip + ":" + s.Config.ExtServInfo.Port + "/Portal/springApi/api/absences"
		return url, rb
	}
	resp, err := s.doRequest(ctx, requestInfo)
	if err != nil {
		return nil, err
	}
	respData := &absenceReason{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
		err = errors.New("The json contract between application and external server is violated: " + err.Error())
		return nil, err
	}
	return respData, nil
}
