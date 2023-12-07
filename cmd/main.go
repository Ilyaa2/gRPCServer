package main

import (
	"gRPCServer/internal/app"
	"log"
)

const configDir = "configs"

func main() {
	err := app.Run(configDir)
	if err != nil {
		log.Fatal(err)
	}
}
