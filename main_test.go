package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
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
