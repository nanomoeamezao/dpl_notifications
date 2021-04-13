package main

import (
	"time"

	"github.com/go-redis/redis/v8"
)

type DirectMessage struct {
	uid     int
	message string
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	redis *redis.Client
	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	//send message to a client with id
	sendToId chan *DirectMessage
}

func newHub(redis *redis.Client) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		sendToId:   make(chan *DirectMessage),
		redis:      redis,
	}
}

func test_spam_direct(h *Hub) {
	for {
		h.sendToId <- &DirectMessage{uid: 111, message: "xdlmao"}
		time.Sleep(5 * time.Second)
	}
}

func test_spam(h *Hub) {
	for {
		h.broadcast <- []byte(time.Now().Format("3:4:5") + " xd lmao")
		time.Sleep(5 * time.Second)
	}
}

func (h *Hub) run() {
	go test_spam_direct(h)
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			go h.handleRedisForClient(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case directMessage := <-h.sendToId:
			for client := range h.clients {
				if client.id == directMessage.uid {
					select {
					case client.send <- []byte(directMessage.message):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
func (h *Hub) handleRedisForClient(client *Client) {
	ticker := time.NewTicker(time.Second * 5)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			func() {
				val, err := h.redis.XRead(ctx, &redis.XReadArgs{
					Streams: []string{"111", client.lastMsgId},
				}).Result()
				if err != nil {
					panic(err)
				}
				if err != redis.Nil {
					for _, stream := range val[0].Messages {
						client.lastMsgId = stream.ID
						h.sendToId <- &DirectMessage{uid: 111, message: stream.Values["msg"].(string)}
					}
				}
			}()
		}
	}
}
