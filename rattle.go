package rattle

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/net/websocket"
)

var (
	c = &Controllers{make(map[string]interface{})}
	// conn  *websocket.Conn
	Debug bool
)

//Controllers type
type Controllers struct {
	C map[string]interface{}
}

//SetControllers bind controllers
func SetControllers(cons ...interface{}) http.Handler {
	for _, con := range cons {
		name := getControllerName(con)
		c.C[name] = con
	}
	return websocket.Handler(WSHandler)
}

func getControllerName(c interface{}) string {
	constring := reflect.ValueOf(c).String()
	f := regexp.MustCompile("\\.([A-Za-z0-9]+) ").FindString(constring)
	f = strings.Trim(f, ". ")
	return f
}

//WSHandler is handler for websocket connections
func WSHandler(ws *websocket.Conn) {
	scanner := bufio.NewScanner(ws)

	for scanner.Scan() && ws != nil {
		go Request(ws, scanner.Bytes())
	}
}

// type Connection struct {
// 	WS *websocket.Conn
// }

func Request(ws *websocket.Conn, bmsg []byte) {
	msg, err := Parsemsg(bmsg)
	if err != nil {
		if Debug {
			log.Println(err, "incoming msg:", string(bmsg))
		}
		return
	}
	msg.WS = ws

	answer, err := msg.Call()
	if err != nil {
		if Debug {
			log.Println(err, "incoming msg:", string(bmsg))
		}
		return
	}
	if answer != nil {
		if err := answer.Send(); err != nil {
			if Debug {
				log.Println(err, "incoming msg:", string(bmsg))
			}
			ws.Close()
		}
	}
}

//Message type: fields From and To RPCMethod type and Data []byte type with payload in json format
type Message struct {
	WS   *websocket.Conn
	From []byte
	To   []byte
	Data []byte
}

//RPCMethod contains name of controller and method
type RPCMethod struct {
	Controller string
	Method     string
}

//NewRPCMethod simple wrapper returned RCPMethod from strings
func NewRPCMethod(controller, method string) RPCMethod {
	return RPCMethod{controller, method}
}

//Parsemsg parse []byte message to type Message
func Parsemsg(msg []byte) (*Message, error) {
	splitted := bytes.SplitN(msg, []byte(" "), 3)
	if len(splitted) != 3 {
		return nil, errors.New("failed incoming message")
	}

	r := new(Message)
	r.From = splitted[0]
	r.To = splitted[1]
	r.Data = splitted[2]

	return r, nil
}

//splitRPC function split string with controller and method to RPCMethod
func splitRPC(rpc []byte) (RPCMethod, error) {
	var r RPCMethod

	splitted := bytes.SplitN(rpc, []byte("."), 2)
	if len(splitted) != 2 {
		return r, errors.New("failed split rpc request")
	}

	r.Controller = strings.Title(string(splitted[0]))
	r.Method = strings.Title(string(splitted[1]))

	return r, nil
}

//Join rpc to one []byte line
func (rpc *RPCMethod) Join() []byte {
	return []byte(fmt.Sprintf("%s.%s", rpc.Controller, rpc.Method))
}

//Call method by name
func (r *Message) Call() (*Message, error) {
	var icontoller interface{}
	var ok bool

	rpc, err := splitRPC(r.To)
	if err != nil {
		return nil, err
	}

	if icontoller, ok = c.C[rpc.Controller]; !ok {
		return nil, errors.New("404 page not found")
	}

	controller := reflect.ValueOf(icontoller)
	method := controller.MethodByName(rpc.Method)
	if !method.IsValid() {
		return nil, errors.New("404 page not found")
	}

	if err := json.Unmarshal(r.Data, controller.Interface()); err != nil {
		return nil, errors.New("failed parse json: " + err.Error())
	}

	//call controller method
	refAnswer := method.Call([]reflect.Value{reflect.ValueOf(r)})
	if len(refAnswer) == 0 || refAnswer[0].Interface() == nil {
		return nil, nil
	}

	a := refAnswer[0].Interface().(*Message)
	if a == nil {
		return nil, nil
	}
	a.From = rpc.Join()

	return a, nil
}

//Bytes convert Message type to []byte, for write to socket
func (r *Message) Bytes() []byte {
	var msg [][]byte

	msg = append(msg, r.From)
	msg = append(msg, r.To)
	msg = append(msg, r.Data)
	msg = append(msg, []byte("\n"))

	return bytes.Join(msg, []byte(" "))
}

func (r *Message) NewMessage(to string, data ...[]byte) *Message {
	msg := &Message{}
	msg.WS = r.WS

	msg.From = getFrom()
	msg.To = []byte(to)

	if len(data) > 0 {
		msg.Data = data[0]
	} else {
		msg.Data = []byte(`{}`)
	}
	return msg
}

func (r *Message) Send() error {
	_, err := r.WS.Write(r.Bytes())
	return err
}

func getFrom() []byte {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[1])

	s := strings.Split(f.Name(), "/")
	funcname := s[len(s)-1]

	fs := strings.Split(funcname, ".")
	if len(fs) != 3 {
		panic(errors.New("failed get the caller function: " + f.Name()))
	}

	controller := regexp.MustCompile("[\\(\\)\\*]").ReplaceAllString(fs[1], "")

	return []byte(fmt.Sprintf("%s.%s", controller, fs[2]))
}
