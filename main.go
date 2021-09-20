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
	guuid "github.com/google/uuid"
	"github.com/gorilla/websocket"
)

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
	// redisResult := sendToStream(rdb, msg)
	// log.Printf("%s", redisResult)
	redisErr := sendToPUBSUB(rdb, msg)
	if redisErr != nil {
		log.Println(redisErr)
	}
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
	upgrader.CheckOrigin = func(r *http.Request) bool {
		log.Println(r.Host)
		if r.Host == "localhost:8080" || r.Host == "localhost" {
			return true
		}
		return false
	}
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
	fcmTokenCookie, err := r.Cookie("fcm")
	if err != nil {
		log.Println(err)
		return
	}
	id := UIDcookie.Value
	uid, _ := strconv.Atoi(id)
	uuid := guuid.NewString()
	lastMsgId := lastMsgCookie.Value
	fcmToken := fcmTokenCookie.Value
	log.Printf("new connect")
	client := &Client{hub: hub, conn: conn, send: make(chan *Message, 256), id: uid, lastMsgId: lastMsgId, control: make(chan bool), uuid: uuid, fcm: fcmToken}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

var ctx = context.Background()

func main() {
	fmt.Printf("listening")
	rdb := initRedis()
	hub := newHub(rdb)
	go hub.run()
	handleRoutes(hub, rdb)
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	err := http.ListenAndServe(fmt.Sprint(":", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func handleRoutes(hub *Hub, rdb *redis.Client) {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		handleAPIRequest(w, r, rdb)
	})

}
