package main

import (
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

func sendToStream(rdb *redis.Client, msg JSONMessage) error {
	args := makeStreamArgs(msg.Params.Id, msg.Params.Msg)
	redisResult := rdb.XAdd(ctx, args).Err()
	return redisResult
}

func makeStreamArgs(id int64, msg string) *redis.XAddArgs {
	return &redis.XAddArgs{
		Stream: strconv.FormatInt(id, 10),
		Values: []interface{}{"msg", msg},
	}

}

func sendToPUBSUB(rdb *redis.Client, msg JSONMessage) error {
	id := strconv.FormatInt(msg.Params.Id, 10)
	message := msg.Params.Msg
	err := rdb.Publish(ctx, id, message).Err()
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

func (h *Hub) handleRedisForClient(client *Client) {
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
		}
	} else if err == redis.Nil {
		log.Print("no new msgs for ", client.id)
	} else {
		log.Print("error for ", client.id, " : ", err)
	}
}
func (h *Hub) subForClient(client *Client) {
	log.Println("subbing: ", client.id)
	channel := fmt.Sprint(client.id)
	sub := h.redis.Subscribe(ctx, channel)
	ch := sub.Channel()

	for {
		select {
		case message := <-ch:
			client.send <- &Message{Id: "-1", Message: message.Payload}

		}
	}

}
