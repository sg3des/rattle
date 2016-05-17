package wsrpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"golang.org/x/net/websocket"
)

var (
	c *Controllers
)

//Controllers type
type Controllers struct {
	C map[string]interface{}
}

//NewControllers bind controllers
func NewControllers(cons map[string]interface{}) http.Handler {
	c = &Controllers{C: cons}
	return websocket.Handler(WSHandler)
}

//WSHandler is handler for websocket connections
func WSHandler(ws *websocket.Conn) {
	scanner := bufio.NewScanner(ws)

	for scanner.Scan() {
		msg := scanner.Bytes()

		r, err := Parsemsg(msg)
		if err != nil {
			log.Println(err)
			continue
		}
		answer, err := r.Call()
		if err != nil {
			log.Println(err)
			continue
		}
		if answer != nil {
			ws.Write(answer.Bytes())
		}
	}
}

//Message type: fields From and To RPCMethod type and Data []byte type with payload in json format
type Message struct {
	From RPCMethod
	To   RPCMethod
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
	var err error

	r := new(Message)
	r.From, err = splitRPC(splitted[0])
	if err != nil {
		return nil, err
	}
	r.To, err = splitRPC(splitted[1])
	if err != nil {
		return nil, err
	}
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

	r.Controller = string(splitted[0])
	r.Method = string(splitted[1])

	return r, nil
}

//Call method by name
func (r *Message) Call() (*Message, error) {
	var icontoller interface{}
	var ok bool

	if icontoller, ok = c.C[r.To.Controller]; !ok {
		return nil, errors.New("404 page not found")
	}

	controller := reflect.ValueOf(icontoller)
	method := controller.MethodByName(r.To.Method)
	if !method.IsValid() {
		return nil, errors.New("404 page not found")
	}

	// data := controller.Elem()
	// jsonStruct := controller.Interface() //data.Interface()
	if err := json.Unmarshal(r.Data, controller.Interface()); err != nil {
		return nil, errors.New("failed parse json: " + err.Error())
	}

	//call controller method
	refAnswer := method.Call([]reflect.Value{reflect.ValueOf(r)})
	if len(refAnswer) == 0 || refAnswer[0].Interface() == nil {
		return nil, nil
	}
	a := refAnswer[0].Interface().(*Message)
	a.From = RPCMethod{r.To.Controller, r.To.Method}

	return a, nil
}

//Bytes convert Message type to []byte, for write to socket
func (r *Message) Bytes() []byte {
	var b []byte

	b = append(b, []byte(fmt.Sprintf("%s.%s ", r.From.Controller, r.From.Method))...)
	b = append(b, []byte(fmt.Sprintf("%s.%s ", r.To.Controller, r.To.Method))...)

	b = append(b, r.Data...)
	b = append(b, '\n')
	return b
}
