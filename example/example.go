package main

import (
	"log"
	"net/http"
	"time"

	"github.com/sg3des/rattle"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func main() {
	if err := server(); err != nil {
		log.Fatal(err)
	}
}

func server() error {
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/rattle/", http.StripPrefix("/rattle/", http.FileServer(http.Dir("../"))))

	//set debug mode
	rattle.Debug = true

	//bind controllers and get handler
	wshandle := rattle.SetControllers(&Main{})
	http.Handle("/ws", wshandle)

	return http.ListenAndServe("127.0.0.1:8080", nil)
}

//Main controller, into fields parse JSON requests
type Main struct {
	Name string
	Text string
}

//Index is method of controller Main on request takes incoming message and possible returne answer message
func (c *Main) Index(r *rattle.Message) *rattle.Message {

	//just send message - insert data to h1 field
	r.NewMessage("#h1", []byte(`Main Index`)).Send()

	//returned message call test.Recieve Frontend function with JSON data
	return r.NewMessage("test.Recieve", []byte(`{"newJSONkey":"`+c.Text+`"}`))
}

//Timer is periodic send data
func (c *Main) Timer(r *rattle.Message) {
	for {
		time.Sleep(time.Second)
		t := time.Now().Local().Format("2006.01.02 15:04:05")
		err := r.NewMessage("#timer", []byte(t)).Send()
		if err != nil {
			return
		}
	}
}

func (c *Main) Something(r *rattle.Message) *rattle.Message {
	tpldata := []byte(`<a href='https://github.com/sg3des/rattle'>rattle link</a>`)

	r.NewMessage("construct.Recieve").Send()

	return r.NewMessage("+body", tpldata)
}
