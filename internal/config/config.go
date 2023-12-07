package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gRPCServer/internal/domain"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

const defaultConfigName = "config.json"

type Config struct {
	ExtServInfo ExternalServerInfo `json:"externalServerInfo"`
	AppServInfo AppServerInfo      `json:"appServerInfo"`
}

type AppServerInfo struct {
	ServerIp          string `json:"serverIp"`
	ServerPort        string `json:"serverPort"`
	LogLevel          string `json:"logLevel"`
	AmountOfWorkers   int    `json:"amountOfWorkers"`
	QueueSize         int    `json:"queueSize"`
	TTLOfItemsInCache int64  `json:"ttlOfItemsInCache"`
}

type ExternalServerInfo struct {
	EmployeeUrlPath string `json:"employeeUrlPath"`
	AbsenceUrlPath  string `json:"absenceUrlPath"`
	Login           string `json:"login"`
	Password        string `json:"password"`
	//millis
	RequestTimeout int `json:"requestTimeout"`
}

// int as input expected
func checkMinInt(minThreshold int) func(interface{}) error {
	return func(value interface{}) error {
		number, ok := value.(int)
		if !ok {
			return errors.New("must be an int")
		}
		if number < minThreshold {
			return errors.New("must be more than: " + strconv.Itoa(minThreshold))
		}
		return nil
	}
}

// int 64 as input expected
func checkMinInt64(minThreshold int) func(interface{}) error {
	return func(value interface{}) error {
		number, ok := value.(int64)
		if !ok {
			return errors.New("must be an int64")
		}
		if number < int64(minThreshold) {
			return errors.New("must be more than: " + strconv.Itoa(minThreshold))
		}
		return nil
	}
}

func validateConfig(config *Config) error {
	err := validation.ValidateStruct(&config.AppServInfo,
		validation.Field(&config.AppServInfo.ServerIp, validation.Required, is.IP),
		validation.Field(&config.AppServInfo.ServerPort, validation.Required, is.Port),
		validation.Field(&config.AppServInfo.AmountOfWorkers,
			validation.Required, validation.By(checkMinInt(1))),
		validation.Field(&config.AppServInfo.QueueSize, validation.Required,
			validation.By(checkMinInt(1))),
		validation.Field(&config.AppServInfo.TTLOfItemsInCache, validation.Required,
			validation.By(checkMinInt64(1))),
		validation.Field(&config.AppServInfo.LogLevel, validation.Required, is.Alpha),
	)
	if err != nil {
		return err
	}
	err = validation.ValidateStruct(&config.ExtServInfo,
		validation.Field(&config.ExtServInfo.EmployeeUrlPath, validation.Required, is.RequestURL),
		validation.Field(&config.ExtServInfo.AbsenceUrlPath, validation.Required, is.RequestURL),
		validation.Field(&config.ExtServInfo.RequestTimeout, validation.Required,
			validation.By(checkMinInt(50))),
		validation.Field(&config.ExtServInfo.Login, validation.Required),
		validation.Field(&config.ExtServInfo.Password, validation.Required),
	)

	return err
}

// ParseJsonConfig parse the config file. Json format expected. Filename could be specified via system variable "config-file-name".
// The config file must be in configs directory. Also there are some validation rules that could be seen in doc directory.
func ParseJsonConfig(configDir string) (*Config, error) {
	var configName string
	flag.StringVar(&configName, "config-file-name", defaultConfigName, "config file name in json format")
	p := filepath.Clean(path.Join(configDir, configName))
	file, err := os.Open(p)
	if err != nil {
		err = fmt.Errorf("config: %w. Details: %v", domain.ErrFileNotExists, err)
		_ = file.Close()
		return nil, err
	}
	config := Config{}

	if err = json.NewDecoder(file).Decode(&config); err != nil {
		err = fmt.Errorf("config: %w. Details: %v", domain.ErrInvalidFileStructure, err)
		_ = file.Close()
		return nil, err
	}
	if err = validateConfig(&config); err != nil {
		err = fmt.Errorf("config: %w. Details: %v", domain.ErrIncorrectValue, err)
		_ = file.Close()
		return nil, err
	}
	_ = file.Close()
	return &config, nil
}
