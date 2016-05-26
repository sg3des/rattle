package rattle

import (
	"bytes"
	"fmt"
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

	// go fakeReciever()

	msg := &Message{To: []byte("FakeController.FakeMethod"), Data: []byte(`{"Name":"testname"}`)}
	if _, err := conn.Write(msg.Bytes()); err != nil {
		t.Error(err)
	}
}

// func fakeReciever() {
// 	scanner := bufio.NewScanner(conn)
// 	for scanner.Scan() {
// 		bmsg := scanner.Bytes()
// 		fmt.Println("msg for frontend:", string(bmsg))
// 	}
// }

func TestParsemsg(t *testing.T) {
	incorrectmsgs := []string{"\n", "{}\n", " test.toMethod\n"}

	for _, smsg := range incorrectmsgs {
		_, err := parsemsg([]byte(smsg))
		if err == nil {
			t.Error("failed parse msg: '" + smsg + "' must be error")
		}
	}

	correctmsg := []byte("FakeController.FakeMethod {\"name\":\"value\"}\n")

	msg, err := parsemsg(correctmsg)
	if err != nil {
		t.Error(err)
	}

	rpc, err := splitRPC(msg.To)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal([]byte(fmt.Sprintf("%s.%s", rpc.Controller, rpc.Method)), msg.To) {
		t.Error("failed inverse transformation controller and method")
	}

	if !bytes.Equal(msg.Bytes(), correctmsg) {
		t.Error("failed convert msg to bytes", string(msg.Bytes()), string(correctmsg))
	}

	c := new(Conn)
	// Conn.NewMessage(to, ...)
	newmsg := c.NewMessage("test.To")
	if !bytes.Equal(newmsg.To, []byte(`test.To`)) {
		t.Error("failed create new message field To fill incorrect")
	}

	if len(newmsg.Data) != 0 {
		t.Error("failed create new message field Data fill incorrect")
	}
}

//****
//Benchmarks
func BenchmarkJSONRequests(b *testing.B) {
	conn, err = websocket.Dial("ws://"+addr+"/ws", "", "http://"+addr)
	if err != nil {
		b.Error(err)
	}

	msg := &Message{To: []byte("FakeController.FakeEmptyMethod"), Data: []byte(`{"Name":"TestValue"}`)}
	bmsg := msg.Bytes()

	for i := 0; i < b.N; i++ {
		conn.Write(bmsg)
	}
}

func BenchmarkEmptyRequests(b *testing.B) {
	conn, err = websocket.Dial("ws://"+addr+"/ws", "", "http://"+addr)
	if err != nil {
		b.Error(err)
	}

	msg := &Message{To: []byte("FakeController.FakeEmptyMethod")}
	bmsg := msg.Bytes()

	for i := 0; i < b.N; i++ {
		conn.Write(bmsg)
	}
}
