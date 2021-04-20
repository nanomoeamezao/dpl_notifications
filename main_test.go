package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestHandleJSONMsg_jrpc(t *testing.T) {
	testmsg := bytes.NewReader([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": {"msg": "23", "id": 23}, "id": 1}`))
	var want = JSONParams{Id: 23, Msg: "23"}
	got, err := decodeJSONMessage(ioutil.NopCloser(testmsg))
	if fmt.Sprintln(want) != fmt.Sprintln(got.Params) && err == nil {
		t.Fatal(got.Params, want)
	}
}

func TestCheckMessage_positive(t *testing.T) {
	testmsg := bytes.NewReader([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": {"msg": "23", "id": 23}, "id": 1}`))
	got, _ := decodeJSONMessage(ioutil.NopCloser(testmsg))
	err := checkDecodedMessage(got)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func TestUpdateClientsLastMessageRedis_positive(t *testing.T) {
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	hub := newHub(rdb)
	client := &Client{hub: hub, conn: nil, send: make(chan *Message, 256), id: 111, lastMsgId: "14888888", control: make(chan bool)}
	cmd, err := updateClientsLastMessageRedis(client, rdb, ctx)
	if err != nil || cmd != "OK" {
		t.Fatal(err.Error(), cmd)
	}
}

func TestReadClientsLastMessageRedis_positive(t *testing.T) {
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	hub := newHub(rdb)
	client := &Client{hub: hub, conn: nil, send: make(chan *Message, 256), id: 111, lastMsgId: "8888", control: make(chan bool)}
	updateClientsLastMessageRedis(client, rdb, ctx)
	res, err := readClientsLastMessageRedis(client, rdb, ctx)
	if err != nil || res != "8888" {
		t.Error(err.Error())
	}
}

func TestSerializeMessages_positive(t *testing.T) {
	messageArray := []*Message{}
	messageArray = append(messageArray, &Message{Id: "11", Message: "kek"})
	messageArray = append(messageArray, &Message{Id: "12", Message: "kek"})
	messageArray = append(messageArray, &Message{Id: "13", Message: "kek"})
	encoded_single, err_single := serializeMessages(messageArray[:1])
	want_single := "[{\"Id\":\"11\",\"Message\":\"kek\"}]"

	encoded_multiple, err_multiple := serializeMessages(messageArray)
	want_multiple := "[{\"Id\":\"11\",\"Message\":\"kek\"},{\"Id\":\"12\",\"Message\":\"kek\"},{\"Id\":\"13\",\"Message\":\"kek\"}]"
	if err_single != nil {
		t.Error(err_single.Error(), " single error")
	} else if encoded_single != want_single {
		t.Error("encoded != want: ", encoded_single)
	}

	if err_multiple != nil {
		t.Error(err_multiple.Error(), " multiple error")
	} else if encoded_multiple != want_multiple {
		t.Error("encoded != want: multiple ", encoded_multiple, " != ", want_multiple)
	}
}
