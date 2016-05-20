package main

import (
	"fmt"
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

	rattle.Debug = true
	ws := rattle.SetControllers(&Main{})
	http.Handle("/ws", ws)

	return http.ListenAndServe("127.0.0.1:8080", nil)
}

type Main struct {
	Name string
	Text string
}

func (c *Main) Index(r *rattle.Message) *rattle.Message {
	fmt.Println(c)
	r.NewMessage("#h1", []byte(`Main Index`)).Send()

	return r.NewMessage("test.Recieve", []byte(`{"newJSONkey":"`+c.Text+`"}`))
}

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
