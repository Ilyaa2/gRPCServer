package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gRPCServer/internal/config"
	"gRPCServer/internal/domain"
	"gRPCServer/pkg/util"
	"net/http"
	"time"
)

type EmployeeRepo struct {
	compositeLogger domain.CompositeLogger
	cfg             *config.ExternalServerInfo
}

func NewEmployeeRepo(cfg *config.ExternalServerInfo, logger domain.CompositeLogger) *EmployeeRepo {
	return &EmployeeRepo{cfg: cfg, compositeLogger: logger}
}

func (e *EmployeeRepo) GetByEmail(ctx context.Context, email string) (*domain.EmployeeData, error) {
	reqID := util.GetReqIDFromContext(ctx)
	requestInfo := func() (string, interface{}) {
		t := time.Now()
		dateFrom := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		dateTo := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
		reqBody := struct {
			DateFrom time.Time `json:"dateFrom"`
			DateTo   time.Time `json:"dateTo"`
			Email    string    `json:"email"`
		}{
			Email:    email,
			DateFrom: dateFrom,
			DateTo:   dateTo,
		}
		return e.cfg.EmployeeUrlPath, reqBody
	}
	resp, err := e.doRequest(ctx, requestInfo)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	respData := &domain.EmployeeData{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil {
		err = fmt.Errorf("%w. Details: %v", domain.ErrViolatedJsonContract, err)
		e.compositeLogger.ApplicationLogger.Error(
			"unexpected error during unmarshalling json",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "repository",
				"function": "doRequest",
				"err":      err,
			})
		return nil, err
	}
	return respData, nil
}

func (e *EmployeeRepo) GetAbsenceReason(ctx context.Context, empData *domain.EmployeeData) (*domain.AbsenceReason, error) {
	reqID := util.GetReqIDFromContext(ctx)
	if empData == nil {
		e.compositeLogger.ApplicationLogger.Error(
			"unexpected error during unmarshalling json",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "repository",
				"function": "GetAbsenceReason",
				"err":      domain.ErrNilData,
			})
		return nil, domain.ErrNilData
	}
	requestInfo := func() (string, interface{}) {
		t := time.Now()
		dateFrom := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		dateTo := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())

		reqBody := struct {
			DateFrom  time.Time `json:"dateFrom"`
			DateTo    time.Time `json:"dateTo"`
			PersonIds []int     `json:"personIds"`
		}{
			//todo добавил только один id
			PersonIds: []int{empData.Data[0].Id},
			DateFrom:  dateFrom,
			DateTo:    dateTo,
		}
		return e.cfg.AbsenceUrlPath, reqBody
	}
	resp, err := e.doRequest(ctx, requestInfo)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	respData := &domain.AbsenceReason{}
	err = json.NewDecoder(resp.Body).Decode(respData)
	if err != nil || respData.Status != "OK" {
		err = fmt.Errorf("%w. Details: %v", domain.ErrViolatedJsonContract, err)
		//TODO ПОВТОР
		e.compositeLogger.ApplicationLogger.Error(
			"unexpected error during unmarshalling json",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "repository",
				"function": "doRequest",
				"err":      err,
			})
		return nil, err
	}
	return respData, nil
}

func (e *EmployeeRepo) doRequest(ctx context.Context, requestInfo func() (string, interface{})) (*http.Response, error) {
	reqID := util.GetReqIDFromContext(ctx)
	select {
	case <-ctx.Done():
		e.compositeLogger.RequestResponseLogger.Warn(
			"context was expired or canceled",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "repository",
				"function": "doRequest",
				"err":      ctx.Err(),
			},
		)
		return nil, ctx.Err()
	default:
		url, requestBody := requestInfo()
		req, cancelCtx, err := e.createRequest(ctx, reqID, url, requestBody)
		defer cancelCtx()
		if err != nil {
			return nil, err
		}
		client := http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			err = fmt.Errorf("%w. Details: %v", domain.ErrExternalServer, err)
			e.compositeLogger.HttpTrafficLogger.Error(
				"error when requesting to external server",
				map[string]interface{}{
					"req-id":   reqID,
					"package":  "repository",
					"function": "doRequest",
					"url":      url,
					"err":      err,
				})
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("%w: status code of response = %v. Status: %v. Details: %v",
				domain.ErrExternalServer, resp.StatusCode, resp.Status, err)
			e.compositeLogger.HttpTrafficLogger.Error(
				"status code of the request is not ok",
				map[string]interface{}{
					"req-id":   reqID,
					"package":  "repository",
					"function": "doRequest",
					"url":      url,
					"err":      err,
				})
			return nil, err
		}

		e.compositeLogger.HttpTrafficLogger.Debug(
			"get response from external server",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "repository",
				"function": "doRequest",
				"response": resp.Body,
				"url":      url,
			},
		)

		return resp, nil
	}
}

func (e *EmployeeRepo) createRequest(ctx context.Context, reqID string, url string,
	requestBody interface{}) (*http.Request, context.CancelFunc, error) {
	jsonReq, err := json.Marshal(requestBody)
	if err != nil {
		err = fmt.Errorf("error during the conversion into JSON: %w. "+
			"Details: %v", domain.ErrInternalServer, err)
		e.compositeLogger.ApplicationLogger.Error(
			"unexpected json conversion error",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "repository",
				"function": "doRequest",
				"err":      err,
			},
		)
		return nil, nil, err
	}
	childCtx, cancel := context.WithTimeout(ctx, time.Duration(e.cfg.RequestTimeout)*time.Millisecond)

	req, err := http.NewRequestWithContext(childCtx, http.MethodPost, url, bytes.NewBuffer(jsonReq))
	if err != nil {
		err = fmt.Errorf("error during the creation of request to send it to an external server: %w. "+
			"Details: %v", domain.ErrInternalServer, err)

		e.compositeLogger.ApplicationLogger.Error(
			"unexpected creation request error",
			map[string]interface{}{
				"req-id":   reqID,
				"package":  "repository",
				"function": "doRequest",
				"err":      err,
			})
		cancel()
		return nil, nil, err
	}

	e.compositeLogger.HttpTrafficLogger.Debug(
		"send request to external server",
		map[string]interface{}{
			"req-id":   reqID,
			"package":  "repository",
			"function": "doRequest",
			"request":  requestBody,
			"url":      url,
		})

	req.SetBasicAuth(e.cfg.Login, e.cfg.Password)
	return req, cancel, nil
}
