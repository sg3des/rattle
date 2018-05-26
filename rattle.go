package rattle

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/sg3des/bytetree"
	"golang.org/x/net/websocket"
)

var Debug bool

type Rattle struct {
	//Connections map contains all available web socket connections
	Connections *bytetree.Tree

	//Routes contains name and interface link on controller
	Routes *bytetree.Tree

	//events
	onConnect    Handler
	onDisconnect Handler
}

//NewRattle initialize new rattle instance
func NewRattle() *Rattle {
	r := &Rattle{
		Connections: bytetree.NewTree(),
		Routes:      bytetree.NewTree(),
	}

	return r
}

//Handler method type
type Handler func(*Request)

//AddRoute handler method by name
func (r *Rattle) AddRoute(name string, h Handler) {
	r.Routes.GrowLeaf([]byte(name), h)
}

//SetOnConnect bind event onconnect
func (r *Rattle) SetOnConnect(f Handler) {
	r.onConnect = f
}

//SetOnDisconnect bind event onDisconnect
func (r *Rattle) SetOnDisconnect(f Handler) {
	r.onDisconnect = f
}

//Handler return http.Handler with websocket
func (r *Rattle) Handler() http.Handler {
	return websocket.Handler(r.wshandler)
}

//wshandler is handler for websocket connections
func (r *Rattle) wshandler(ws *websocket.Conn) {
	c := r.newСonnection(ws)
	if Debug {
		log.Println("new connection:", string(c.key))
	}

	scanner := bufio.NewScanner(ws)
	for scanner.Scan() && c != nil {
		r.call(c, scanner.Bytes())
	}

	r.Disconnect(c)
}

func (r *Rattle) call(c *Connection, data []byte) {
	req, err := r.parseRequest(data)
	if err != nil {
		if Debug {
			log.Println(err)
		}
		return
	}

	req.conn = c

	if Debug {
		log.Printf("request: %s %s", req.To, req.Data)
	}

	switch req.Type {
	case "stream":
		err = r.stream(req)
	default:
		err = r.request(req)
	}

	if err != nil && Debug {
		log.Println(err)
	}
}

//Connection is main struct contains websocket.Conn
type Connection struct {
	key []byte
	w   *bytes.Buffer
	ws  *websocket.Conn
}

func (r *Rattle) newСonnection(ws *websocket.Conn) *Connection {
	key := []byte(fmt.Sprintf("%p", ws))
	if v, ok := r.Connections.LookupLeaf(key); ok {

		return v.(*Connection)
	}

	c := &Connection{
		key: key,
		ws:  ws,
		w:   bytes.NewBuffer([]byte{}),
	}

	r.Connections.GrowLeaf(key, c)

	if r.onConnect != nil {
		go r.onConnect(&Request{conn: c})
	}

	return c
}

//Disconnect end the current connection
func (r *Rattle) Disconnect(c *Connection) {
	if c != nil {
		if r.onDisconnect != nil {
			r.onDisconnect(&Request{conn: c})
		}

		r.Connections.CutLeaf(c.key)
		c.ws.Close()
		c = nil
	}
}

type Request struct {
	conn *Connection

	To     string
	Type   string
	URL    string `json:"url"`
	Data   json.RawMessage
	Stream json.RawMessage

	File *File
}

//File structure
type File struct {
	Name string `json:"name"`
	Size int    `json:"size"`

	Buffer *bytes.Buffer

	SliceSize int `json:"slicesize"`
}

func (req *Request) DecodeTo(v interface{}) error {
	return json.Unmarshal(req.Data, v)
}

//parsemsg parse []byte message to type Message
func (r *Rattle) parseRequest(buf []byte) (req *Request, err error) {
	err = json.Unmarshal(buf, &req)
	return
}

//Request is main function, takes connection and raw incoming message
func (r *Rattle) request(req *Request) error {
	v, ok := r.Routes.LookupLeaf([]byte(req.To))
	if !ok {
		return errors.New("404 page not found")
	}

	v.(Handler)(req)
	return nil
}

func (r *Rattle) stream(req *Request) error {
	req.conn.ws.Write([]byte("stream --"))

	err := json.Unmarshal(req.Stream, &req.File)
	if err != nil {
		if Debug {
			log.Println("rattle: failed unmarshal file header:", err)
		}
		return err
	}

	req.File.Buffer = bytes.NewBuffer([]byte{})
	log.Println(req.File)

STREAM:
	for {
		line := make([]byte, req.File.SliceSize)
		n, err := req.conn.ws.Read(line)
		if err != nil {
			if Debug {
				log.Println(err)
			}
			return err

		}
		line = line[:n]

		rr, err := r.parseRequest(line)
		// if err != nil {
		// 	if Debug {
		// 		log.Println(err)
		// 	}
		// 	continue
		// }

		if err == nil {
			switch rr.Type {
			case "chunk":
				req.conn.ws.Write([]byte("stream --"))
				continue STREAM
			case "finish":
				break STREAM
			}
		}

		if !bytes.Equal(line, []byte("\n")) {
			req.File.Buffer.Write(line)
		}
		// }

		if err != nil {
			break STREAM
		}
	}

	return r.request(req)
}

//Message type:
//  To - contains name of called function, must be filled!
//  Data may contains payload in json format - for backend, or json,html or another for frontend, not necessary.
type Message struct {
	conn *Connection
	To   []byte
	Data []byte
}

//NewMessage create answer message
func (req *Request) NewMessage(to string, data []byte) *Message {
	msg := &Message{
		To:   []byte(to),
		conn: req.conn,
	}

	if len(data) > 0 {
		msg.Data = data
	}

	return msg
}

//Bytes convert Message type to []byte, for write to socket
func (m *Message) Bytes() (bmsg []byte) {
	buf := bytes.NewBuffer(bmsg)

	buf.Write(m.To)
	buf.WriteRune(' ')
	buf.Write(m.Data)

	bmsg = buf.Bytes()
	// bmsg = regexp.MustCompile("(?m)\n+$").ReplaceAll(bmsg, []byte("\n"))

	return
}

//Send message to connection
func (m *Message) Send() error {
	if m.conn == nil || m.conn.ws == nil {
		err := errors.New("connection is nil")
		if Debug {
			log.Println(err)
		}
		return err
	}
	_, err := m.conn.ws.Write(m.Bytes())

	if Debug {
		log.Printf("send message to: %s with data: %s", m.To, m.Data)
	}

	return err
}

//Broadcast send one message for all available connections(users)
func (r *Rattle) Broadcast(m *Message) {
	for _, v := range r.Connections.PickAllLeafs() {
		v.(Connection).ws.Write(m.Bytes())
	}
}
