package main

import "time"

type DirectMessage struct {
	uid     int
	message string
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	//send message to a client with id
	sendToId chan *DirectMessage
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		sendToId:   make(chan *DirectMessage),
	}
}

func test_spam(h *Hub) {
	for {
		h.broadcast <- []byte(time.Now().Format("3:4:5") + " xd lmao")
		time.Sleep(5 * time.Second)
	}
}

func (h *Hub) run() {
	go test_spam(h)
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
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
