package main

import (
	"context"
	"fmt"
	"gRPCServer/internal/transport/grpc/sources/dataModification"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	address = "127.0.0.1:8080"
	n       = 100
)

func main() {
	mu := &sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(p int, mu *sync.Mutex) {
			defer wg.Done()
			time.Sleep(time.Nanosecond)
			request(p, mu)
		}(i, mu)
	}

	wg.Wait()
	fmt.Println("Done it")
}

func request(p int, mutex *sync.Mutex) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	defer conn.Close()
	c := dataModification.NewPersonalInfoClient(conn)

	cd := &dataModification.ContactDetails{
		DisplayName: "Иванов Семен Петрович",
		Email:       "example@gmail.com",
		MobilePhone: strconv.Itoa(p),
		WorkPhone:   strconv.Itoa(p),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	resp, err := c.GetReasonOfAbsence(ctx, cd)
	//mutex.Lock()
	if err != nil {
		log.Fatalf("some error: %v", err)
	}
	log.Print(
		"my data was: \n",
		cd.String(), "\n",
		"the data I get", "\n",
		resp.String(),
	)

	//mutex.Unlock()
}
