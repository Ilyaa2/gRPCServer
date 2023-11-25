package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path"
)

const defaultConfigName = "config.json"

type Config struct {
	AppServInfo AppServerInfo      `json:"appServerInfo"`
	ExtServInfo ExternalServerInfo `json:"externalServerInfo"`
}

type AppServerInfo struct {
	ServerIp        string `json:"serverIp"`
	ServerPort      string `json:"serverPort"`
	AmountOfWorkers int    `json:"amountOfWorkers"`
	LogLevel        string `json:"logLevel"`
}

type ExternalServerInfo struct {
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// todo ОШИБКА ПРИ КОТОРОЙ JSON СТРУКТУРА НЕПРАВИЛЬНАЯ ДОЛЖНА БЫТЬ КАСТОМНАЯ И БОЛЕЕ ИНФОРМАТИВНАЯ
func ParseJsonConfig(configDir string) (*Config, error) {
	var configName string
	flag.StringVar(&configName, "config-file-name", defaultConfigName, "config file name in json format")
	file, err := os.Open(path.Join(configDir, configName))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	config := Config{}

	if err = json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal(err)
		return nil, err
	}
	if err = validateConfig(&config); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return &config, nil
}

// TODO можно указать отрицательное количество воркеров
func validateConfig(config *Config) error {
	return nil
}
