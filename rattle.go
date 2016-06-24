package rattle

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
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

	if Debug {
		log.SetFlags(log.Lshortfile)
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
		inmsg, err := c.parsemsg(scanner.Bytes())
		if err != nil {
			if Debug {
				log.Println(err)
			}
			continue
		}
		switch inmsg.Type {
		case "json":

			c.request(inmsg) //removed prefix go - sadly, but multi-threaded has a lot of problems. if at the same time cause the same method incoming data will mix
		case "stream":
			c.stream(inmsg)
		}
	}

	c.Disconnect()
}

type File struct {
	Name string `json:"name"`
	Size int    `json:"size"`

	Buffer *bytes.Buffer

	SliceSize int `json:"slicesize"`
}

func (c *Conn) stream(inmsg *Inmsg) {
	c.WS.Write([]byte("stream --"))

	err := json.Unmarshal(inmsg.Stream, &c.File)
	if err != nil {
		if Debug {
			log.Println("rattle: failed unmarshal file header: " + err.Error())
		}
		return
	}

	c.File.Buffer = bytes.NewBuffer([]byte{})

STREAM:
	for {
		line := make([]byte, c.File.SliceSize)
		n, err := c.WS.Read(line)
		line = line[:n]
		// log.Println(string(line[:100]))

		inmsg, _ := c.parsemsg(line)
		// if err != nil {
		// 	if Debug {
		// 		log.Println(err)
		// 	}
		// 	continue
		// }
		if inmsg != nil {
			switch inmsg.Type {
			case "chunk":
				c.WS.Write([]byte("stream --"))
				continue STREAM
			case "finish":
				break STREAM
			}
		} //else {

		if !bytes.Equal(line, []byte("\n")) {
			c.File.Buffer.Write(line)
		}
		// }

		if err != nil {
			break STREAM
		}
	}

	go c.request(inmsg)
}

type Inmsg struct {
	To     string
	Type   string
	JSON   json.RawMessage
	Stream json.RawMessage
}

//parsemsg parse []byte message to type Message
func (c *Conn) parsemsg(bmsg []byte) (msg *Inmsg, err error) {
	err = json.Unmarshal(bmsg, &msg)
	return
}

func GetBoundary() []byte {
	return []byte("--" + multipart.NewWriter(bytes.NewBuffer([]byte{})).Boundary()[:16] + "--")
}

//newConnection
func newConnection(ws *websocket.Conn) *Conn {
	c := &Conn{WS: ws, W: bytes.NewBuffer([]byte{}), boundary: GetBoundary()}

	Connections[c] = true
	if onConnect != nil {
		onConnect(c)
	}

	return c
}

//Disconnect end the current connection
func (c *Conn) Disconnect() {
	if c != nil {
		c.WS.Close()

		if onDisconnect != nil {
			onDisconnect(c)
		}

		delete(Connections, c)
		c = nil
	}
}

//Conn is main struct contains websocket.Conn
type Conn struct {
	W    *bytes.Buffer
	WS   *websocket.Conn
	File *File
	Raw  []byte

	boundary []byte
}

//Request is main function, takes connection and raw incoming message
//  1) parse message
//  2) call specified method of controller
//  3) and if a answer is returned, then write it to connection
func (c *Conn) request(inmsg *Inmsg) {
	// fmt.Println(string(bmsg))

	m, err := c.call(inmsg)
	if err != nil {
		if Debug {
			log.Println(err, "get msg to:", inmsg.To, "with data: ", inmsg.JSON)
		}
		return
	}
	if m != nil {
		if err := m.Send(); err != nil {
			if Debug {
				log.Println(err, "send msg to:", string(m.To), "with data: ", string(m.Data))
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
func splitRPC(funcname string) (*methodRPC, error) {
	r := &methodRPC{}

	splitted := strings.SplitN(funcname, ".", 2)
	// splitted := bytes.SplitN(rpc, []byte("."), 2)
	if len(splitted) != 2 {
		return r, errors.New("failed split rpc request")
	}

	r.Controller = strings.Title(splitted[0])
	r.Method = strings.Title(splitted[1])

	return r, nil
}

//Call method by name
func (c *Conn) call(m *Inmsg) (*Message, error) {
	rpc, err := splitRPC(m.To)
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

	clearStruct(controller) // otherwise the new data fall over the previous

	if len(m.JSON) > 1 {
		json.Unmarshal(m.JSON, &conInterface)
		c.Raw = m.JSON
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

func clearStruct(controller reflect.Value) {
	//clear struct
	for i := 0; i < controller.Elem().NumField(); i++ {
		switch controller.Elem().Field(i).Type().String() {
		case "string":
			controller.Elem().Field(i).Set(reflect.ValueOf(""))
		case "int":
			controller.Elem().Field(i).Set(reflect.ValueOf(0))
		case "bool":
			controller.Elem().Field(i).Set(reflect.ValueOf(false))
		}
	}
}

//NewMessage create answer message
func (c *Conn) NewMessage(to string, data ...[]byte) *Message {
	msg := &Message{To: []byte(to), conn: c}
	// msg.conn = c

	if len(data) > 0 {
		msg.Data = data[0]
	} else if c.W.Len() > 0 {
		msg.Data = c.W.Bytes()
		c.W.Truncate(0)
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

//Bytes convert Message type to []byte, for write to socket
func (m *Message) Bytes() (bmsg []byte) {
	buf := bytes.NewBuffer(bmsg)

	buf.Write(m.To)
	buf.WriteRune(' ')
	buf.Write(m.Data)

	bmsg = buf.Bytes()
	bmsg = regexp.MustCompile("\n+$").ReplaceAll(bmsg, []byte("\n"))

	return
}

//Send message to connection
func (m *Message) Send() error {
	if m.conn == nil || m.conn.WS == nil {
		return errors.New("connection is nil")
	}
	_, err := m.conn.WS.Write(m.Bytes())
	return err
}

//Broadcast send one message for all available connections(users)
func Broadcast(m *Message) {
	for conn := range Connections {
		if conn != nil && conn.WS != nil {
			conn.WS.Write(m.Bytes())
		}
	}
}
