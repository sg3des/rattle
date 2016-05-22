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
	//Debug mode enable error output
	Debug bool
	//Connections map contains all available web socket connections
	Connections = make(map[*websocket.Conn]bool)
	//Controllers contains name and interface link on controller
	Controllers = make(map[string]interface{})
)

//SetControllers bind controllers
func SetControllers(userContollers ...interface{}) http.Handler {
	for _, c := range userContollers {
		Controllers[getControllerName(c)] = c
	}

	return websocket.Handler(WSHandler)
}

//getControllerName determine name of struct through reflect
func getControllerName(c interface{}) string {
	constring := reflect.ValueOf(c).String()
	f := regexp.MustCompile("\\.([A-Za-z0-9]+) ").FindString(constring)
	return strings.Trim(f, ". ")
}

//WSHandler is handler for websocket connections
func WSHandler(ws *websocket.Conn) {
	Connections[ws] = true
	scanner := bufio.NewScanner(ws)

	for scanner.Scan() && ws != nil {
		go Request(ws, scanner.Bytes())
	}
}

//Request is main function, takes connection and raw incoming message
//  1) parse message
//  2) call specified method of controller
//  3) and if a answer is returned, then write it to connection
func Request(ws *websocket.Conn, bmsg []byte) {
	msg, err := Parsemsg(bmsg)
	if err != nil {
		if Debug {
			log.Println(err, "rattle incoming msg:", string(bmsg))
		}
		return
	}
	msg.WS = ws

	answer, err := msg.Call()
	if err != nil {
		if Debug {
			log.Println(err, "rattle incoming msg:", string(bmsg))
		}
		return
	}
	if answer != nil {
		if err := answer.Send(); err != nil {
			if Debug {
				log.Println(err, "rattle incoming msg:", string(bmsg))
			}

			ws.Close()
			delete(Connections, ws)
		}
	}
}

//Message type:
//  From - name of calling function, autofill, can not be empty.
//  To - contains name of called function, must be filled!
//  Data may contains payload in json format - for backend, or json,html or another for frontend, not necessary.
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
// func NewRPCMethod(controller, method string) RPCMethod {
// 	return RPCMethod{controller, method}
// }

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
	rpc, err := splitRPC(r.To)
	if err != nil {
		return nil, err
	}

	var conInterface interface{}
	var ok bool
	if conInterface, ok = Controllers[rpc.Controller]; !ok {
		return nil, errors.New("404 page not found")
	}

	controller := reflect.ValueOf(conInterface)
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

	if !regexp.MustCompile("\n$").Match(r.Data) {
		r.Data = append(r.Data, []byte("\n")...)
	}
	msg = append(msg, r.Data)

	return bytes.Join(msg, []byte(" "))
}

//NewMessage create answer message
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

//Send message to connection
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
		if Debug {
			log.Println("rattle warning: failed get the caller function: " + f.Name())
		}
		return []byte("empty")
	}

	controller := regexp.MustCompile("[\\(\\)\\*]").ReplaceAllString(fs[1], "")

	return []byte(fmt.Sprintf("%s.%s", controller, fs[2]))
}

//Broadcast send one message for all available connections(users)
func (r *Message) Broadcast() {
	for conn := range Connections {
		if conn != nil {
			conn.Write(r.Bytes())
		}
	}
}
