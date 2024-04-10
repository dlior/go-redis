package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

type Member struct {
	Member string `json:"name"`
	Score  uint   `json:"score"`
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("server is healthy.\n"))
}

func getLeaderboard(w http.ResponseWriter, r *http.Request) {
	var members []Member
	ctx := context.Background()

	players, err := client.ZRevRangeWithScores(ctx, leaderboard, 0, -1).Result()
	if err != nil {
		log.Println("failed to read leadersboard.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, player := range players {
		member := Member{player.Member.(string), uint(player.Score)}
		members = append(members, member)
	}

	err = json.NewEncoder(w).Encode(members)
	if err != nil {
		log.Println("failed to encode members.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTopN(w http.ResponseWriter, r *http.Request) {
	var members []Member
	ctx := context.Background()

	n := r.URL.Query().Get("n")
	num, _ := strconv.Atoi(n)

	players, err := client.ZRevRangeWithScores(ctx, leaderboard, 0, int64(num-1)).Result()
	if err != nil {
		log.Println("failed to read leadersboard sorted set.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, player := range players {
		member := Member{player.Member.(string), uint(player.Score)}
		members = append(members, member)
	}

	err = json.NewEncoder(w).Encode(members)
	if err != nil {
		log.Println("failed to encode members.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func addPlayer(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	ctx := context.Background()

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println("failed to decode payload.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	player := payload["name"]
	exists, err := client.SIsMember(ctx, players, player).Result()
	if err != nil {
		log.Println("failed to query player.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch exists {
	case true:
		log.Println("player", player, "already exists.")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, player, "already exists.")
	case false:
		err := client.SAdd(ctx, players, player).Err()
		if err != nil {
			log.Println("failed to add player.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func play(w http.ResponseWriter, r *http.Request) {
	go func() {
		ctx := context.Background()

		for {
			log.Println("game simulation started...")

			players, err := client.SMembers(ctx, players).Result()
			if err != nil {
				log.Println("failed to read players set.")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, player := range players {
				err := client.ZIncrBy(ctx, leaderboard, float64(rand.IntN(20)+1), player).Err()
				if err != nil {
					log.Println("failed to update leaderboard sorted set.")
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	w.WriteHeader(http.StatusOK)
}
