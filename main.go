package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

func serveTestpage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	cookie, err := r.Cookie("UID")
	if err != nil {
		log.Println(err)
		return
	}
	id := cookie.Value
	uid, _ := strconv.Atoi(id)
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), id: uid, lastMsgId: "0"}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
}

var ctx = context.Background()

func main() {
	fmt.Printf("listening")
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:3333",
		Password: "",
		DB:       0,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		fmt.Printf("error connecting to redis: %s", err)
		panic("redis error")
	}

	hub := newHub(rdb)
	go hub.run()
	http.HandleFunc("/", serveTestpage)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
