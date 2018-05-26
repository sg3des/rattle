package rattle

import (
	"log"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

var (
	addr = "127.0.0.1:8088"

	r   *Rattle
	err error
)

func init() {
	Debug = true
	log.SetFlags(log.Lshortfile)
}

func TestNewRattle(t *testing.T) {
	r = NewRattle()
	r.AddRoute("echo", echo)
	r.AddRoute("empty", empty)
	http.Handle("/ws", r.Handler())

	go func() {
		http.ListenAndServe(addr, nil)
	}()

	time.Sleep(500 * time.Millisecond)
}

func echo(r *Request) {
	r.NewMessage("echo", r.Data)
}

func empty(r *Request) {

}

func TestRequest(t *testing.T) {
	conn, err := websocket.Dial("ws://"+addr+"/ws", "", "http://"+addr)
	if err != nil {
		t.Error(err)
	}

	if _, err := conn.Write([]byte(`{"to":"echo","type":"data","data":"some data"}` + "\n")); err != nil {
		t.Error(err)
	}

	time.Sleep(500 * time.Millisecond)
}

//More tests are needed

//Benchmarks
func BenchmarkRequests(b *testing.B) {
	b.StopTimer()
	Debug = false
	conn, err := websocket.Dial("ws://"+addr+"/ws", "", "http://"+addr)
	if err != nil {
		b.Error(err)
	}

	msg := []byte([]byte(`{"to":"empty"}` + "\n"))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		conn.Write(msg)
	}
}
