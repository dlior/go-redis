package handlers

import (
	"go-redis/logger"
	"net/http"
	"time"
)

func SayHello(w http.ResponseWriter, r *http.Request) {
	logger.LogCh <- logger.LogEntry{Time: time.Now(), Severity: logger.LogDebug, Message: "hello"}
	w.Write([]byte("hello"))
}
