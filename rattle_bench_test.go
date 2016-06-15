package rattle

import (
	"testing"

	"golang.org/x/net/websocket"
)

func BenchmarkJSONRequests(b *testing.B) {
	conn, err := websocket.Dial("ws://"+addr+"/ws", "", "http://"+addr)
	if err != nil {
		b.Error(err)
	}

	msg := &Message{To: []byte("FakeController.FakeEmptyMethod"), Data: []byte(`{"Name":"TestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValueTestValue"}`)}
	bmsg := msg.Bytes()

	for i := 0; i < b.N; i++ {
		conn.Write(bmsg)
	}
}

func BenchmarkEmptyRequests(b *testing.B) {
	conn, err := websocket.Dial("ws://"+addr+"/ws", "", "http://"+addr)
	if err != nil {
		b.Error(err)
	}

	msg := &Message{To: []byte("FakeController.FakeEmptyMethod")}
	bmsg := msg.Bytes()

	for i := 0; i < b.N; i++ {
		conn.Write(bmsg)
	}
}
