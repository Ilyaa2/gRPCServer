package main

import (
	"context"
	"fmt"
	dm "gRPCServer/internal/rpc/dataModification"
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
	c := dm.NewPersonalInfoClient(conn)

	cd := &dm.ContactDetails{
		Email: "" + strconv.Itoa(p),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	r, err := c.GetReasonOfAbsence(ctx, cd)
	//mutex.Lock()
	if err != nil {
		log.Fatalf("some error: %v", err)
	}
	log.Printf("AnswerName - %s, The number was - %d", r.DisplayName, p)
	//mutex.Unlock()
}
