package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Hub struct {
	// Registered clients.
	clients map[int]*Client

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
		clients:    make(map[int]*Client),
		redis:      redis,
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

func (h *Hub) run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
			go test_spam_direct(client)
			lastMessage, err := readClientsLastMessageRedis(client, h.redis, ctx)
			if err == nil {
				client.lastMsgId = lastMessage
			}
			log.Printf("registering %d", client.id)
			go h.handleRedisForClient(client, ctx)

		case client := <-h.unregister:
			if _, ok := h.clients[client.id]; ok {
				log.Printf("unregistering %d", client.id)
				delete(h.clients, client.id)
				client.control <- true
				close(client.send)
				close(client.control)
			}
			// case message := <-h.broadcast:
			// 	for _, client := range h.clients {
			// 		select {
			// 		case client.send <- message:
			// 		default:
			// 			close(client.send)
			// 			delete(h.clients, client.id)
			// 		}
			// 	}
		}
	}
}
func (h *Hub) handleRedisForClient(client *Client, ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			h.readRedisMessages(client, client.lastMsgId)
			updateClientsLastMessageRedis(client, h.redis, ctx)
		case <-client.control:
			return
		}
	}
}

func (h *Hub) readRedisMessages(client *Client, startID string) {
	log.Printf("reading redis, last msg: %s", startID)
	val, err := h.redis.XRead(ctx, &redis.XReadArgs{
		Streams: []string{strconv.Itoa(client.id), startID},
		Block:   5 * time.Millisecond, //FUCKING WHY?????????????????
	}).Result()
	if err != nil {
		log.Println(err)
	}
	if err != redis.Nil {
		for _, stream := range val[0].Messages {
			client.lastMsgId = stream.ID
			client.send <- &Message{Id: stream.ID, Message: stream.Values["msg"].(string)}
		}
	}
}

func updateClientsLastMessageRedis(client *Client, redis *redis.Client, ctx context.Context) (string, error) {
	lastMessageId := client.lastMsgId
	cmd, err := redis.Set(ctx, fmt.Sprintf("%dlast", client.id), lastMessageId, 0).Result()
	return cmd, err
}

func readClientsLastMessageRedis(client *Client, redis *redis.Client, ctx context.Context) (string, error) {
	res, err := redis.Get(ctx, fmt.Sprint(client.id, "last")).Result()
	return res, err
}
