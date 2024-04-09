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

const usersSet = "users"
const gameLeaderboard = "leaderboard"

func init() {
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("failed to connect to redis. error message - %v", err)
	}

	log.Println("successfully connected to redis")

	err = client.Del(context.Background(), usersSet).Err()
	if err != nil {
		log.Println("could not delete set", usersSet, err)
	}

	err = client.Del(context.Background(), gameLeaderboard).Err()
	if err != nil {
		log.Println("could not delete sorted set", gameLeaderboard, err)
	}

	for i := 1; i <= 10; i++ {
		err = client.SAdd(context.Background(), usersSet, "user-"+strconv.Itoa(i)).Err()
		if err != nil {
			log.Println("could not add user to set", err)
		}
	}

	log.Println("added users to set")
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", addUser).Methods(http.MethodPost)
	r.HandleFunc("/play", play).Methods(http.MethodGet)
	r.HandleFunc("/top/{n}", leaderboard).Methods(http.MethodGet)

	log.Println("started HTTP server....")
	log.Fatal(http.ListenAndServe(":8080", r))
}
