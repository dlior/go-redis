package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-redis/handlers"
	"go-redis/logger"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func init() {
	ctx := context.Background()
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.LogCh <- logger.LogEntry{Time: time.Now(), Severity: logger.LogError, Message: fmt.Sprintf("%s.", err)}
		return
	}

	logger.LogCh <- logger.LogEntry{Time: time.Now(), Severity: logger.LogInfo, Message: "successfully connected to redis."}
}

func main() {
	go logger.Logger()

	r := http.NewServeMux()

	r.HandleFunc("GET /", handlers.SayHello)

	http.ListenAndServe(":8080", r)

	logger.DoneCh <- struct{}{}
}
