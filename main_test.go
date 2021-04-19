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
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, 256), id: 111, lastMsgId: "14888888", control: make(chan bool)}
	cmd, err := updateClientsLastMessageRedis(client, rdb, ctx)
	fmt.Print("test redis ", cmd, "\n")
	if err != nil && cmd != "OK" {
		t.Fatal(err.Error(), cmd)
	}
}
