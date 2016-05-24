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
		imsg := scanner.Bytes()
		bmsg := make([]byte, len(imsg))
		copy(bmsg, imsg)

		go Request(ws, bmsg)
	}
}

//Request is main function, takes connection and raw incoming message
//  1) parse message
//  2) call specified method of controller
//  3) and if a answer is returned, then write it to connection
func Request(ws *websocket.Conn, bmsg []byte) {
	msg, err := parsemsg(bmsg)
	if err != nil {
		if Debug {
			log.Fatalln(err, "msg:", string(bmsg))
		}
		return
	}
	msg.WS = ws

	a, err := msg.Call()
	if err != nil {
		if Debug {
			log.Fatalln(err, "msg:", msg)
		}
		return
	}
	if a != nil {
		if err := a.Send(); err != nil {
			if Debug {
				log.Fatalln(err, "msg:", bmsg)
			}

			ws.Close()
			delete(Connections, ws)
		}
	}
}

//Message type:
//  To - contains name of called function, must be filled!
//  Data may contains payload in json format - for backend, or json,html or another for frontend, not necessary.
type Message struct {
	WS   *websocket.Conn
	To   []byte
	Data []byte
}

//methodRPC contains name of controller and method
type methodRPC struct {
	Controller string
	Method     string
}

var reg = regexp.MustCompile("(?i)^[a-z0-9]+\\.[a-z0-9]+$")

//parsemsg parse []byte message to type Message
func parsemsg(bmsg []byte) (*Message, error) {
	splitted := bytes.SplitN(bmsg, []byte(" "), 2)
	if len(splitted) == 0 {
		return nil, errors.New("failed incoming message")
	}

	r := &Message{To: bytes.Trim(splitted[0], " \n\r")}
	if !reg.Match(r.To) {
		return r, errors.New("incoming message contains invalid characters")
	}

	if len(splitted) == 2 {
		r.Data = bytes.Trim(splitted[1], " \n\r")
	}

	return r, nil
}

//splitRPC function split string with controller and method to RPCMethod
func splitRPC(rpc []byte) (*methodRPC, error) {
	r := &methodRPC{}

	splitted := bytes.SplitN(rpc, []byte("."), 2)
	if len(splitted) != 2 {
		return r, errors.New("failed split rpc request")
	}

	r.Controller = strings.Title(string(splitted[0]))
	r.Method = strings.Title(string(splitted[1]))

	return r, nil
}

//Join rpc to one []byte line
func (rpc *methodRPC) Join() []byte {
	return []byte(fmt.Sprintf("%s.%s", rpc.Controller, rpc.Method))
}

//Call method by name
func (r *Message) Call() (*Message, error) {
	a := &Message{}
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

	if len(r.Data) > 1 {
		// fmt.Println(string(r.Data))
		if err := json.Unmarshal(r.Data, controller.Interface()); err != nil {
			return nil, errors.New("failed parse json: " + err.Error())
		}
	}

	//call controller method
	refAnswer := method.Call([]reflect.Value{reflect.ValueOf(r)})
	if len(refAnswer) == 0 || refAnswer[0].Interface() == nil {
		return nil, nil
	}

	a = refAnswer[0].Interface().(*Message)
	if a == nil {
		return nil, nil
	}

	return a, nil
}

//Bytes convert Message type to []byte, for write to socket
func (r *Message) Bytes() (bmsg []byte) {
	buf := bytes.NewBuffer(bmsg)

	buf.Write(r.To)
	buf.WriteRune(' ')
	buf.Write(r.Data)

	bmsg = buf.Bytes()
	bmsg = regexp.MustCompile("\n*$").ReplaceAll(bmsg, []byte("\n"))

	return
}

//NewMessage create answer message
func (r *Message) NewMessage(to string, data ...[]byte) *Message {
	msg := &Message{WS: r.WS, To: []byte(to)}

	if len(data) > 0 {
		msg.Data = data[0]
	}

	return msg
}

//Send message to connection
func (r *Message) Send() error {
	_, err := r.WS.Write(r.Bytes())
	return err
}

//Broadcast send one message for all available connections(users)
func (r *Message) Broadcast() {
	for conn := range Connections {
		if conn != nil {
			conn.Write(r.Bytes())
		}
	}
}
