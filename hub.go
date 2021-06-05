package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"firebase.google.com/go/messaging"
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
	//fcm
	fcm *messaging.Client
}

func newHub(redis *redis.Client, fcm *messaging.Client) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		redis:      redis,
		fcm:        fcm,
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
			h.clients[client.uuid] = client
			// go test_spam_direct(client)
			log.Printf("registering %d", client.id)
			go h.handleRedisForClient(client, ctx)

		case client := <-h.unregister:
			if _, ok := h.clients[client.uuid]; ok {
				log.Printf("unregistering %d", client.id)
				delete(h.clients, client.uuid)
				client.control <- true
				close(client.send)
				close(client.control)
			}
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
	if err != redis.Nil && err == nil {
		for _, stream := range val[0].Messages {
			client.lastMsgId = stream.ID
			message := &Message{Id: stream.ID, Message: stream.Values["msg"].(string)}
			client.send <- message
			h.sendToFcm(client, message)
		}
	} else if err == redis.Nil {
		log.Print("no new msgs for ", client.id)
	} else {
		log.Print("error for ", client.id, " : ", err)
	}
}

func (h *Hub) sendToFcm(client *Client, message *Message) {
	fcmMessage := &messaging.Message{
		Data: map[string]string{
			"message": message.Message,
		},
		Notification: &messaging.Notification{
			Title: "ConfLab",
			Body:  message.Message,
		},
		Token: client.fcm,
	}
	resp, err := h.fcm.Send(ctx, fcmMessage)
	if err != nil {
		log.Println("fcm send error: ", err)
	}
	log.Println("fcm response: ", resp)
}
