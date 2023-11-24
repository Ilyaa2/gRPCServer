package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

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
func ParseJsonConfig() *Config {
	var configPath string
	flag.StringVar(&configPath, "config-path", "config/config.json", "config file path in json format")
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	config := Config{}
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	return &config
}

// TODO можно указать отрицательное количество воркеров
func validateConfig() error {
	return nil
}
