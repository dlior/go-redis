package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var client *redis.Client

const playerPrefix = "player-"
const players = "players"
const leaderboard = "leaderboard"

func init() {
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis.\nerror message - %v", err)
		return
	}

	if err := client.Del(ctx, players).Err(); err != nil {
		log.Printf("failed to delete players set.\nerror message - %v", err)
		return
	}

	if err := client.Del(ctx, leaderboard).Err(); err != nil {
		log.Printf("failed to delete leaderboard sorted set.\nerror message - %v", err)
		return
	}

	for i := 1; i <= 10; i++ {
		if err := client.SAdd(ctx, players, playerPrefix+strconv.Itoa(i)).Err(); err != nil {
			log.Printf("failed to add player to players set.\nerror message - %v", err)
			return
		}
	}

	log.Println("successfully connected to redis.")
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", health).Methods(http.MethodGet)
	r.HandleFunc("/add-player", addPlayer).Methods(http.MethodPost)
	r.HandleFunc("/play", play).Methods(http.MethodPost)
	r.HandleFunc("/top/{n}", top).Methods(http.MethodGet)
	r.HandleFunc("/leaders", leaders).Methods(http.MethodGet)

	log.Println("HTTP server is running on port 8080.")
	log.Fatalln(http.ListenAndServe(":8080", r))
}
