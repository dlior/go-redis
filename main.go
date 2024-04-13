package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var client *redis.Client
var pubsub *redis.PubSub
var Users map[string]*websocket.Conn
var upgrader = websocket.Upgrader{}

const chatChannel = "chats"

func init() {
	Users = map[string]*websocket.Conn{}
}

func main() {
	ctx := context.Background()
	client = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal("ping failed. could not connect", err)
	}
	startChatBroadcaster()

	http.HandleFunc("/chat/", chat)
	server := http.Server{Addr: ":8080", Handler: nil}

	go func() {
		fmt.Println("started server")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server", err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)
	<-exit

	fmt.Println("exit signalled")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//clean up all connected connections
	for _, conn := range Users {
		conn.Close()
	}

	pubsub.Unsubscribe(context.Background(), chatChannel)
	pubsub.Close()

	server.Shutdown(ctx)

	fmt.Println("application shut down")
}

func chat(w http.ResponseWriter, r *http.Request) {
	user := strings.TrimPrefix(r.URL.Path, "/chat/")

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	Users[user] = c
	fmt.Println(user, "in chat")

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			//fmt.Println("read message error", err)
			_, ok := err.(*websocket.CloseError)
			if ok {
				fmt.Println("connection closed by", user)
				err := c.Close()
				if err != nil {
					fmt.Println("error closing ws connection", err)
				}
				delete(Users, user)
				fmt.Println("closed websocket connection and removed user session")
			}
			break
		}
		err = client.Publish(context.Background(), chatChannel, user+":"+string(message)).Err()
		if err != nil {
			fmt.Println("publish failed", err)
		}
	}
}

func startChatBroadcaster() {
	go func() {
		fmt.Println("listening to messages")
		pubsub = client.Subscribe(context.Background(), chatChannel)
		messages := pubsub.Channel()
		for message := range messages {
			from := strings.Split(message.Payload, ":")[0]
			//broadcast to all
			for user, peer := range Users {
				if from != user {
					peer.WriteMessage(websocket.TextMessage, []byte(message.Payload))
				}
			}
		}
	}()
}
