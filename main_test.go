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
	got, err := decodeApiJSONMessage(ioutil.NopCloser(testmsg))
	if fmt.Sprintln(want) != fmt.Sprintln(got.Params) && err == nil {
		t.Fatal(got.Params, want)
	}
}

func TestCheckMessage_positive(t *testing.T) {
	testmsg := bytes.NewReader([]byte(`{"jsonrpc": "2.0", "method": "subtract", "params": {"msg": "23", "id": 23}, "id": 1}`))
	got, _ := decodeApiJSONMessage(ioutil.NopCloser(testmsg))
	err := checkDecodedMessage(got)
	if err != nil {
		t.Fatalf("%s", err)
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

func TestJsonPUBSUB_positive(t *testing.T) {
	marshalled := `{"Id":"11","Message":"test"}`
	result, err := unmarshalPUBSUB(marshalled)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if result.Id != "11" || result.Message != "test" {
		t.Error("incorrect results: ", result.Id, result.Message)
	}
}

func TestJsonPUBSUB_pairwise(t *testing.T) {
	marshalledArray := []string{`{"Id":"11","Message":"test"}`, `{"Id":"11"}`, `{"Message":"test"}`, `{}`, `{"Id":"11", "Message":11}`, `{"Id":11, "Message":"test"}`}
	// both good, 1 missing x 2 , 2 missing , 1 not string x 2
	for i, v := range marshalledArray {
		res, err := unmarshalPUBSUB(v)
		if err != nil && i == 0 {
			t.Error("unmarshal failed: ", err)
		} else if err != nil {

		} else if res.Id == "" || res.Message == "" {
			t.Error("missing args: ", res)
		}

	}

}
