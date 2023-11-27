package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	"net/http"
	"strconv"
	"time"
)

const (
	EmployeeUrlPath = "/Portal/springApi/api/employees"
	AbsenceUrlPath  = "/Portal/springApi/api/absences"
)

type EmployeeRepo struct {
	cfg *config.ExternalServerInfo
}

func NewEmployeeRepo(cfg *config.ExternalServerInfo) *EmployeeRepo {
	return &EmployeeRepo{cfg: cfg}
}

// todo ПРОСМОТРЕТЬ ПО ВСЕМУ ПРОЕКТУ, ЧТОБ ВЕЗДЕ БЫЛИ ЗАКРЫТЫ СОЕДИНЕНИЯ.
// todo так же чекай текущий контекст
// todo установи таймаут на соединение с сервером. Если таймаут истек - верни ошибку.
// todo тут или в doRequest нужно закрывать resp или resp.Body (метод close)
func (e *EmployeeRepo) GetByEmail(ctx context.Context, email string) (*domain.EmployeeData, error) {
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
		url := "https://" + e.cfg.Ip + ":" + e.cfg.Port + EmployeeUrlPath
		return url, rb
	}
	resp, err := e.doRequest(ctx, requestInfo)
	if err != nil {
		err = errors.New("External server error" + err.Error())
		return nil, err
	}
	respData := &domain.EmployeeData{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil || respData.Status != "OK" {
		err = errors.New("The json contract between application and external server is violated or there was incorrect data in the request" + err.Error())
		return nil, err
	}
	return respData, nil
}

// todo возможно, если пользователю нужно возвращать id, то нужно создать новую сущность, а не только contactDetails
// todo хардкод в url -> сделать так чтоб в конфиге это задавалось
// todo так же чекай текущий контекст
// todo []int{empData.Data[0].Id}. Разобраться при первом запросе можно ли послать массив email или нужно много делать запросов по всем email. Задать этот вопрос
func (e *EmployeeRepo) GetAbsenceReason(ctx context.Context, empData *domain.EmployeeData) (*domain.AbsenceReason, error) {
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
		url := "https://" + e.cfg.Ip + ":" + e.cfg.Port + AbsenceUrlPath
		return url, rb
	}
	resp, err := e.doRequest(ctx, requestInfo)
	if err != nil {
		return nil, err
	}
	respData := &domain.AbsenceReason{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil || respData.Status != "OK" {
		err = errors.New("The json contract between application and external server is violated or there was incorrect data in the request" + err.Error())
		return nil, err
	}
	return respData, nil
}

// todo смотреть за контекстом, его могут отменить
// todo при запросе установить дедлайн при котором сервер должен ответить. Дедлайн брать из конфигурации
func (e *EmployeeRepo) doRequest(ctx context.Context, requestInfo func() (string, interface{})) (*http.Response, error) {
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

	req.SetBasicAuth(e.cfg.Login, e.cfg.Password)

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
