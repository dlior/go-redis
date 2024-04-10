package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var client *redis.Client

const playerPrefix = "player-"
const players = "players"
const leaderboard = "leaderboard"

func init() {
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis.\nerror message - %s.", err)
	}

	if err := client.Del(ctx, players).Err(); err != nil {
		log.Fatalf("failed to connect to delete players set.\nerror message - %s.", err)
	}

	if err := client.Del(ctx, leaderboard).Err(); err != nil {
		log.Fatalf("failed to connect to delete leaderboard sorted set.\nerror message - %s.", err)
	}

	for i := 1; i <= 10; i++ {
		err := client.SAdd(ctx, players, playerPrefix+strconv.Itoa(i)).Err()
		if err != nil {
			log.Fatalf("failed to add player to players set.\nerror message - %s.", err)
		}
	}

	log.Println("successfuly connected to redis.")
}

func main() {
	r := http.NewServeMux()

	r.HandleFunc("GET /health", health)
	r.HandleFunc("GET /leaderboard", getLeaderboard)
	r.HandleFunc("GET /top", getTopN)
	r.HandleFunc("POST /add-player", addPlayer)
	r.HandleFunc("POST /play", play)

	log.Println("HTTP server started on port 8080...")
	log.Fatalln(http.ListenAndServe(":8080", r))
}
