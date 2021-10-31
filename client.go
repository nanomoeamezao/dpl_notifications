package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *Message

	// last recieved message from redis
	lastMsgId string

	//client id in conflab
	id int

	//google uuid
	uuid string

	control chan bool
}

type Message struct {
	//message number
	Id string
	//message text
	Message string
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				log.Print("send channel closed")
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.hub.unregister <- c
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Print("WS writer error: ", err.Error())
				c.hub.unregister <- c
				return
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			messageArray := []*Message{}
			messageArray = append(messageArray, message)
			for i := 0; i < n; i++ {
				messageArray = append(messageArray, <-c.send)
			}
			encodedMessages, err := serializeMessages(messageArray)
			if err != nil {
				log.Print("serialisation error: ", err.Error())
				return
			} else {
				w.Write([]byte(encodedMessages))
			}

			if err := w.Close(); err != nil {
				c.hub.unregister <- c
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.unregister <- c
				return
			}
		}
	}
}

func serializeMessages(messages []*Message) (string, error) {
	encodedMessages, err := json.Marshal(messages)
	return string(encodedMessages), err
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("readpump error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		log.Println("recieved message from client ", string(message))

	}

}
