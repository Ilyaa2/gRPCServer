package app

import (
	"context"
	"gRPCServer/internal/config"
	"gRPCServer/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(configDir string) {
	cfg, err := config.ParseJsonConfig(configDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	s := server.NewServer(cfg)

	go func() {
		s.Run()
	}()
	log.Print("server is working")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	s.GracefulStop(ctx)
	log.Print("server was stopped")
}

//1)
//2)graceful stop (grpc) не согласован c пулом. Если grpc ожидает пока все соединения
//ослужатся в течение какого либо времени
