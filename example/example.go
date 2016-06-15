package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/sg3des/rattle"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/rattle/", http.StripPrefix("/rattle/", http.FileServer(http.Dir("../"))))

	//set debug mode
	rattle.Debug = true

	//bind event on Connect
	rattle.SetOnConnect(OnConnect)
	rattle.SetOnDisconnect(OnDisconnect)

	//bind controllers and get handler
	wshandle := rattle.SetControllers(&Main{})
	http.Handle("/ws", wshandle)

	println("web server listen on 127.0.0.1:8080")
	http.ListenAndServe("127.0.0.1:8080", nil)
}

//Main controller, into fields parse JSON requests
//!in real project controller will be located in another package
type Main struct {
	Text string
}

//Index is method of controller Main on request takes incoming message and possible return answer message
func (c *Main) Index(r *rattle.Conn) *rattle.Message {
	//return answer - insert data to field with id description
	return r.NewMessage("=#description", []byte(`Rattle is tiny websocket double-sided RPC framework, designed for create dynamic web applications`))
}

//JSON method
func (c *Main) JSON(r *rattle.Conn) *rattle.Message {
	data, err := json.Marshal(c)
	if err != nil {
		return r.NewMessage("+#errors", []byte("failed parse JSON request, error: "+err.Error()))
	}
	//call "test.RecieveJSON frontend function and send to it JSON data"
	return r.NewMessage("test.RecieveJSON", data)
}

//RAW method
func (c *Main) RAW(r *rattle.Conn) *rattle.Message {
	//call "test.RecieveRAW frontend function and send to raw data"
	return r.NewMessage("test.RecieveRAW", []byte(c.Text))
}

//Timer is example of periodic send data, note the that function does not return anything
func (c *Main) Timer(r *rattle.Conn) {
	for {
		t := time.Now().Local().Format("2006.01.02 15:04:05")
		if err := r.NewMessage("=#timer", []byte(t)).Send(); err != nil {
			//if err then connection is closed
			return
		}

		time.Sleep(time.Second)
	}
}

func (c *Main) File(r *rattle.Conn) {
	log.Println("incoming file:", r.File.Name, "size:", r.File.Buffer.Len())
	ioutil.WriteFile(r.File.Name, r.File.Buffer.Bytes(), 0644)
}

//OnConnect handler function for event connection
func OnConnect(r *rattle.Conn) {
	log.Println("someone is connected")
}

//OnDisconnect handler
func OnDisconnect(r *rattle.Conn) {
	log.Println("someone is disconnected")
}
