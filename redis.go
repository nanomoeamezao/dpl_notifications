package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"strconv"
	"time"
)

var localRedisOpts = &redis.Options{
	Addr: "localhost:6379",
	DB:   0,
}

// TODO: redis as object
func sendToStream(rdb *redis.Client, msg ApiJSONMessage) (string, error) {
	args := makeStreamArgs(msg.Params.Id, msg.Params.Msg)
	redisId, redisResult := rdb.XAdd(ctx, args).Result()
	return redisId, redisResult
}

func makeStreamArgs(id int64, msg string) *redis.XAddArgs {
	return &redis.XAddArgs{
		Stream: strconv.FormatInt(id, 10),
		Values: []interface{}{"msg", msg},
	}

}

func sendToPUBSUB(rdb *redis.Client, msg ApiJSONMessage, msgId string) error {
	id := strconv.FormatInt(msg.Params.Id, 10)
	psMessage := Message{Id: msgId, Message: msg.Params.Msg}
	psJSON, err := json.Marshal(psMessage)
	if err != nil {
		log.Print("marshal error: ", err)
		return err
	}
	err = rdb.Publish(ctx, id, psJSON).Err()
	return err
}

func initRedis() *redis.Client {
	redisUrl := os.Getenv("REDIS_URL")
	fmt.Println(redisUrl)
	var redisOptions = &redis.Options{}
	if redisUrl == "" {
		redisOptions = localRedisOpts
	} else {
		redisOptions, _ = redis.ParseURL(redisUrl)
	}
	rdb := redis.NewClient(redisOptions)
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("error connecting to redis: %s", err)
		panic("redis error")
	}
	return rdb
}

func (h *Hub) readRedisMessages(client *Client, startID string) {
	log.Printf("reading redis, last msg: %s", startID)
	val, err := h.redis.XRead(ctx, &redis.XReadArgs{
		Streams: []string{strconv.Itoa(client.id), startID},
		Block:   5 * time.Millisecond, //mandatory block argument?
	}).Result()
	if err != redis.Nil && err == nil {
		for _, stream := range val[0].Messages {
			client.lastMsgId = stream.ID
			message := &Message{Id: stream.ID, Message: stream.Values["msg"].(string)}
			client.send <- message
		}
	} else if err == redis.Nil {
		log.Print("no new msgs for ", client.id)
	} else {
		log.Print("error for ", client.id, " : ", err)
	}
	return
}
func (h *Hub) subForClient(client *Client) {
	log.Println("subbing: ", client.id)
	channel := fmt.Sprint(client.id)
	sub := h.redis.Subscribe(ctx, channel)
	ch := sub.Channel()
	defer func() {
		err := sub.Unsubscribe(ctx)
		if err != nil {
			log.Print("unsub error")
		}
	}()
	for {
		select {
		case <-client.control:
			return
		case message := <-ch:
			if message != nil {
				decodedMessage, err := unmarshalPUBSUB(message.Payload)
				if err != nil {
					log.Print("bad message: ", err)
				} else {
					client.send <- decodedMessage
				}
			}

		}
	}
}

func unmarshalPUBSUB(encMessage string) (*Message, error) {
	var decodedMessage Message
	err := json.Unmarshal([]byte(encMessage), &decodedMessage)
	if err != nil {
		log.Print("unmarshal error: ", err)
		return nil, err
	}
	return &decodedMessage, nil
}
