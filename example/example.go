package main

import (
	"net/http"
	"time"

	"github.com/sg3des/rattle"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/rattle/", http.StripPrefix("/rattle/", http.FileServer(http.Dir("../"))))

	//set debug mode
	rattle.Debug = true

	//bind controllers and get handler
	wshandle := rattle.SetControllers(&Main{})
	http.Handle("/ws", wshandle)

	println("web server listen on 127.0.0.1:8080")
	http.ListenAndServe("127.0.0.1:8080", nil)
}

//Main controller, into fields parse JSON requests
//!in real project controller will be located in another package
type Main struct {
	Name string
	Text string
}

//Index is method of controller Main on request takes incoming message and possible return answer message
func (c *Main) Index(r *rattle.Message) *rattle.Message {
	//just send message - insert data to h1 field
	r.NewMessage("#description", []byte(`Rattle is tiny websocket double-sided RPC framework, designed for create web applications`)).Send()

	//returned message call test.Recieve Frontend function with JSON data
	return r.NewMessage("test.Recieve", []byte(`{"newJSONkey":"`+c.Text+`"}`))
}

//Timer is periodic send data
func (c *Main) Timer(r *rattle.Message) {
	for {

		t := time.Now().Local().Format("2006.01.02 15:04:05")

		if err := r.NewMessage("#timer", []byte(t)).Send(); err != nil {
			//if err then connection is closed
			return
		}
		time.Sleep(time.Second)
	}
}

//Something is just one more method
func (c *Main) Something(r *rattle.Message) *rattle.Message {
	tpldata := []byte(`this link just appended to the body: <a href='https://github.com/sg3des/rattle'>rattle link</a>`)

	r.NewMessage("construct.Recieve").Send()

	return r.NewMessage("+body", tpldata)
}
