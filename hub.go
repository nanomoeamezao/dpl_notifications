package main

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

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
}

func newHub(redis *redis.Client) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		redis:      redis,
	}
}

func test_spam_direct(c *Client) {
	for {
		c.send <- []byte("xdlmao")
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
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// go test_spam_direct(client)
			log.Printf("registering %d", client.id)
			go h.handleRedisForClient(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("unregistering %d", client.id)
				delete(h.clients, client)
				client.control <- true
				close(client.send)
				close(client.control)
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
			h.readRedisMessages(client)
		case <-client.control:
			return
		}
	}
}

func (h *Hub) readRedisMessages(client *Client) {
	log.Printf("reading redis, last msg: %s", client.lastMsgId)
	val, err := h.redis.XRead(ctx, &redis.XReadArgs{
		Streams: []string{"111", "1618344951278-0"},
		Block:   5 * time.Millisecond, //FUCKING WHY?????????????????
	}).Result()
	if err != nil {
		log.Println(err)
	}
	if err != redis.Nil {
		for _, stream := range val[0].Messages {
			client.lastMsgId = stream.ID
			client.send <- []byte(stream.Values["msg"].(string))
		}
	}
}
