package main

import (
	"gRPCServer/internal/app"
)

const configDir = "configs"

func main() {
	app.Run(configDir)
}
