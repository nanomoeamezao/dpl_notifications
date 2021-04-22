package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

type JSONParams struct {
	Msg string
	Id  int64
}

type JSONMessage struct {
	Jsonrpc string
	Method  string
	Params  JSONParams
	Id      int64
}

func handleAPIRequest(w http.ResponseWriter, r *http.Request, rdb *redis.Client) {
	encodedMessage := r.Body
	msg, err := decodeJSONMessage(encodedMessage)
	if err != nil {
		log.Println("failed to decode message: ", msg)
		w.Write([]byte(fmt.Sprintf(`{"jsonrpc": "2.0", "result":"failed", "error":{"code":-32700, "message":"message not decoded"}, id:"%d"}`, msg.Id)))
		return
	}
	errMessage := checkDecodedMessage(msg)
	if errMessage != nil {
		log.Println("decoded message does not contain needed info", errMessage)
		w.Write([]byte(fmt.Sprintf(`{"jsonrpc": "2.0", "result":"failed", "error":{"code":-32602, "message":"message contained incorrect data"}, id:"%d"}`, msg.Id)))
		return
	}
	redisResult := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: strconv.FormatInt(msg.Params.Id, 10),
		Values: []interface{}{"msg", msg.Params.Msg},
	}).Err()
	log.Printf("%s", redisResult)
	w.Write([]byte(fmt.Sprintf(`{"jsonrpc": "2.0", "result":success, id:"%d"}`, msg.Id)))
}

func decodeJSONMessage(encodedMessage io.ReadCloser) (JSONMessage, error) {
	var message JSONMessage
	encodedNonstream, err := ioutil.ReadAll(encodedMessage)
	if err == nil {
		json.Unmarshal(encodedNonstream, &message)
	} else {
		log.Println(err)
	}
	return message, err
}

func checkDecodedMessage(message JSONMessage) error {
	if message.Params.Id == 0 || message.Params.Msg == "" {
		return errors.New("message integrity check failed")
	} else {
		return nil
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	UIDcookie, err := r.Cookie("UID")
	if err != nil {
		log.Println(err)
		return
	}
	lastMsgCookie, err := r.Cookie("lastID")
	if err != nil {
		log.Println(err)
		return
	}
	id := UIDcookie.Value
	uid, _ := strconv.Atoi(id)
	lastMsgId := lastMsgCookie.Value
	log.Printf("new connect")
	client := &Client{hub: hub, conn: conn, send: make(chan *Message, 256), id: uid, lastMsgId: lastMsgId, control: make(chan bool)}
	client.hub.register <- client

	go client.writePump()
	go client.maintain()
}

var ctx = context.Background()

func main() {
	fmt.Printf("listening")
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		redisUrl = "redis:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisUrl,
		DB:   0,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("error connecting to redis: %s", err)
		panic("redis error")
	}

	hub := newHub(rdb)
	go hub.run(ctx)
	http.HandleFunc("/", serveTestpage)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		handleAPIRequest(w, r, rdb)
	})
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
