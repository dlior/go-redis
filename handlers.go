package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type Member struct {
	Member string `json:"name"`
	Score  uint   `json:"score"`
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("application is healthy.\n"))
}

func addPlayer(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string

	ctx := context.Background()

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("failed to decode payload.\nerror message - %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	player := payload["name"]
	exists, err := client.SIsMember(ctx, players, player).Result()
	if err != nil {
		log.Printf("failed to verify player.\nerror message - %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch exists {
	case true:
		log.Println("player", string(player), "already exists.")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintln(w, string(player), "already exists.")
	case false:
		err := client.SAdd(ctx, players, string(player)).Err()
		if err != nil {
			log.Printf("failed to add player to players set.\nerror message - %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func play(w http.ResponseWriter, r *http.Request) {
	go func() {
		for {
			fmt.Println("simulation started...")

			ctx := context.Background()
			members, err := client.SMembers(ctx, players).Result()
			if err != nil {
				log.Printf("failed to read players set.\nerror message - %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, member := range members {
				err := client.ZIncrBy(ctx, leaderboard, float64(rand.Intn(20)+1), member).Err()
				if err != nil {
					log.Printf("failed to read players set.\nerror message - %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

func leaders(w http.ResponseWriter, r *http.Request) {
	var members []Member
	ctx := context.Background()
	leaders, err := client.ZRevRangeWithScores(ctx, leaderboard, 0, -1).Result()
	if err != nil {
		log.Printf("failed to read leaderboard sorted set.\nerror message - %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, leader := range leaders {
		member := Member{leader.Member.(string), uint(leader.Score)}
		members = append(members, member)
	}

	err = json.NewEncoder(w).Encode(members)
	if err != nil {
		log.Printf("failed to encode leaders.\nerror message - %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func top(w http.ResponseWriter, r *http.Request) {
	var members []Member

	n := r.PathValue("n")
	num, _ := strconv.Atoi(n)

	ctx := context.Background()
	leaders, err := client.ZRevRangeWithScores(ctx, leaderboard, 0, int64(num-1)).Result()
	if err != nil {
		log.Printf("failed to read leaderboard sorted set.\nerror message - %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, leader := range leaders {
		member := Member{leader.Member.(string), uint(leader.Score)}
		members = append(members, member)
	}

	err = json.NewEncoder(w).Encode(members)
	if err != nil {
		log.Printf("failed to encode leaders.\nerror message - %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
