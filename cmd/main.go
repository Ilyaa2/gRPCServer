package main

import (
	"context"
	"gRPCServer/config"
	dm "gRPCServer/internal/rpc/dataModification"
	"gRPCServer/internal/server"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	cfg := config.ParseJsonConfig()
	ctx, cancel := context.WithCancel(context.Background())
	s := &server.Server{
		//todo нужно положить число в канале тоже в конфиг
		AbsenceJobsQueue: make(chan server.AbsenceJob, 10),
		Config:           cfg,
		Ctx:              ctx,
		Cancel:           cancel,
	}
	address := cfg.AppServInfo.ServerIp + ":" + cfg.AppServInfo.ServerPort
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s.SetWorkersPool()
	grpcServ := grpc.NewServer()
	dm.RegisterPersonalInfoServer(grpcServ, s)
	log.Printf("Starting gRPC listener on adress: " + address)
	if err := grpcServ.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
