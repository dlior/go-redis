package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func addUser(w http.ResponseWriter, req *http.Request) {

	userB, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("failed to read payload", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	exists, err := client.SIsMember(context.Background(), usersSet, string(userB)).Result()
	if err != nil {
		log.Println("could not check user", string(userB), "in set", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists {
		err = client.SAdd(context.Background(), usersSet, string(userB)).Err()
		if err != nil {
			log.Println("could not add user", string(userB), "to set", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Println("added user", string(userB))
	} else {
		log.Println("user", string(userB), "already exists")
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintln(w, string(userB)+" already exists")
	}
}

func play(w http.ResponseWriter, req *http.Request) {
	//simulate

	go func() {
		for {
			log.Println("game simulation running...")

			members, err := client.SMembers(context.Background(), usersSet).Result()
			if err != nil {
				log.Println("could get users", err)
				return
			}

			for _, member := range members {
				_, err := client.ZIncrBy(context.Background(), gameLeaderboard, float64(rand.Intn(20)+1), member).Result()
				if err != nil {
					log.Println("could get incr score for member", err)
					return
				}
				//log.Println("updated score for member", member, "current score", currScore)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

func leaderboard(w http.ResponseWriter, req *http.Request) {

	n := mux.Vars(req)["n"]
	log.Println("fetching top", n, "players")

	num, _ := strconv.Atoi(n)

	//top 5
	leaders, err := client.ZRevRangeWithScores(context.Background(), gameLeaderboard, 0, int64(num-1)).Result()

	if err != nil {
		log.Println("failed to query sorted set", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(leaders)
	if err != nil {
		log.Println("failed to encode leaderboard info", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("successfully fetched leaderboard info....")
}
