package main

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Hub struct {
	// Registered clients.
	clients map[string]*Client

	redis *redis.Client
	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	//redis object
	rdb *RDB
}

func newHub(rdb *RDB) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		rdb:        rdb,
	}
}

func test_spam_direct(c *Client) {
	for {
		select {
		case <-c.control:
			log.Print("stopped spam")
			return
		default:
			c.send <- &Message{Id: "-1", Message: "spam"}
			time.Sleep(5 * time.Second)
		}
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.uuid] = client
			// go test_spam_direct(client)
			log.Printf("registering %d", client.id)
			//sync or goroutine
			h.rdb.readRedisMessages(client, client.lastMsgId)
			go h.rdb.subForClient(client)

		case client := <-h.unregister:
			if _, ok := h.clients[client.uuid]; ok {
				log.Printf("unregistering %d", client.id)
				delete(h.clients, client.uuid)
				close(client.send)
				close(client.control)
			}
		}
	}
}
