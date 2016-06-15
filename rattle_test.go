package rattle

import (
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

var (
	conn *websocket.Conn
	err  error
	addr = "127.0.0.1:8088"
)

// ****
// fake controllers for tests
type FakeController struct {
	Name string
}

func (c *FakeController) FakeMethod(r *Conn) *Message {
	// fmt.Println("recieve message:", c)
	r.NewMessage("tovoid0").Send()
	return r.NewMessage("tovoid1")
}

func (c *FakeController) FakeEmptyMethod(r *Conn) {
	time.Sleep(time.Second)
}

func init() {
	Debug = true
	log.SetFlags(log.Lshortfile)

	wshandle := SetControllers(
		&FakeController{},
	)
	http.Handle("/ws", wshandle)

	go func() {
		err = http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
	}()

	//so the server had go up
	time.Sleep(300 * time.Millisecond)
}

// TESTS
func TestSetControllers(t *testing.T) {
	// controllers already set in init, just check the correctness of this

	if len(Controllers) != 1 {
		t.Fatal("failed set controllers, length of Controllers map is incorrect")
	}

	if conInterface, ok := Controllers["FakeController"]; ok {
		controller := reflect.ValueOf(conInterface)
		if !controller.IsValid() {
			t.Error("failed set controllers, incorrect reflect of controller interface")
		}
		if !controller.MethodByName("FakeEmptyMethod").IsValid() {
			t.Error("failed set controllers, required method not found")
		}
		if !controller.MethodByName("FakeMethod").IsValid() {
			t.Error("failed set controllers, required method not found")
		}
	} else {
		t.Error("failed set controllers, incorrect determine name of controller")
	}
}

func TestRequest(t *testing.T) {
	conn, err = websocket.Dial("ws://"+addr+"/ws", "", "http://"+addr)
	if err != nil {
		t.Error(err)
	}

	msg := &Message{To: []byte("FakeController.FakeMethod"), Data: []byte(`{"Name":"testname"}`)}
	if _, err := conn.Write(msg.Bytes()); err != nil {
		t.Error(err)
	}
}

//More tests are needed
