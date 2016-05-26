package rattle

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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
	Connections = make(map[*Conn]bool)
	//Controllers contains name and interface link on controller
	Controllers = make(map[string]interface{})

	//events
	onConnect    func(*Conn)
	onDisconnect func(*Conn)
)

//SetOnConnect bind event onconnect
func SetOnConnect(f func(*Conn)) {
	onConnect = f
}

//SetOnDisconnect bind event onDisconnect
func SetOnDisconnect(f func(*Conn)) {
	onDisconnect = f
}

//SetControllers bind controllers
func SetControllers(userContollers ...interface{}) http.Handler {
	for _, c := range userContollers {
		Controllers[getControllerName(c)] = c
	}

	return websocket.Handler(wshandler)
}

//getControllerName determine name of struct through reflect
func getControllerName(c interface{}) string {
	constring := reflect.ValueOf(c).String()
	f := regexp.MustCompile("\\.([A-Za-z0-9]+) ").FindString(constring)
	return strings.Trim(f, ". ")
}

//wshandler is handler for websocket connections
func wshandler(ws *websocket.Conn) {
	c := newConnection(ws)

	scanner := bufio.NewScanner(ws)
	for scanner.Scan() && c != nil {
		//create new slice, otherwise the incoming data can be changed before request executed
		bmsg := append([]byte{}, scanner.Bytes()...)
		go c.request(bmsg)
	}

	c.Disconnect()
}

//newConnection
func newConnection(ws *websocket.Conn) *Conn {
	c := &Conn{ws}
	Connections[c] = true

	if onConnect != nil {
		onConnect(c)
	}

	return c
}

//Disconnect end the current connection
func (c *Conn) Disconnect() {
	if c != nil {
		c.WebSocket.Close()

		if onDisconnect != nil {
			onDisconnect(c)
		}

		delete(Connections, c)
		c = nil
	}
}

//Conn is main struct contains websocket.Conn
type Conn struct {
	WebSocket *websocket.Conn
}

//Request is main function, takes connection and raw incoming message
//  1) parse message
//  2) call specified method of controller
//  3) and if a answer is returned, then write it to connection
func (c *Conn) request(bmsg []byte) {
	msg, err := parsemsg(bmsg)
	if err != nil {
		if Debug {
			log.Println(err, "msg:", string(bmsg))
		}
		return
	}

	m, err := c.call(msg)
	if err != nil {
		if Debug {
			log.Println(err, "msg:", string(bmsg))
		}
		return
	}
	if m != nil {
		if err := m.Send(); err != nil {
			if Debug {
				log.Println(err, "msg:", string(bmsg))
			}

			// c.Disconnect()
		}
	}
}

//methodRPC contains name of controller and method
type methodRPC struct {
	Controller string
	Method     string
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

// //Join rpc to one []byte line
// func (rpc *methodRPC) join() []byte {
// 	return []byte(fmt.Sprintf("%s.%s", rpc.Controller, rpc.Method))
// }

//Call method by name
func (c *Conn) call(r *Message) (*Message, error) {
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
	refAnswer := method.Call([]reflect.Value{reflect.ValueOf(c)})
	if len(refAnswer) == 0 || refAnswer[0].Interface() == nil {
		return nil, nil
	}

	a := refAnswer[0].Interface().(*Message)
	if a == nil {
		return nil, nil
	}

	return a, nil
}

//NewMessage create answer message
func (c *Conn) NewMessage(to string, data ...[]byte) *Message {
	msg := &Message{To: []byte(to), conn: c}
	// msg.conn = c

	if len(data) > 0 {
		msg.Data = data[0]
	}

	return msg
}

//Message type:
//  To - contains name of called function, must be filled!
//  Data may contains payload in json format - for backend, or json,html or another for frontend, not necessary.
type Message struct {
	conn *Conn
	To   []byte
	Data []byte
}

var reg = regexp.MustCompile("(?i)^[a-z0-9]+\\.[a-z0-9]+$")

//parsemsg parse []byte message to type Message
func parsemsg(bmsg []byte) (*Message, error) {
	splitted := bytes.SplitN(bmsg, []byte(" "), 2)
	if len(splitted) == 0 {
		return nil, errors.New("failed incoming message")
	}

	m := &Message{To: bytes.Trim(splitted[0], " \n\r")}
	if !reg.Match(m.To) {
		return m, errors.New("incoming message contains invalid characters")
	}

	if len(splitted) == 2 {
		m.Data = bytes.Trim(splitted[1], " \n\r")
	}

	return m, nil
}

//Bytes convert Message type to []byte, for write to socket
func (m *Message) Bytes() (bmsg []byte) {
	buf := bytes.NewBuffer(bmsg)

	buf.Write(m.To)
	buf.WriteRune(' ')
	buf.Write(m.Data)

	bmsg = buf.Bytes()
	bmsg = regexp.MustCompile("\n*$").ReplaceAll(bmsg, []byte("\n"))

	return
}

//Send message to connection
func (m *Message) Send() error {
	_, err := m.conn.WebSocket.Write(m.Bytes())
	// if err != nil {
	// 	m.conn.Disconnect()
	// }
	return err
}

//Broadcast send one message for all available connections(users)
func Broadcast(m *Message) {
	for conn := range Connections {
		if conn != nil {
			conn.WebSocket.Write(m.Bytes())
			// conn.Write(m.Bytes())
		}
	}
}
