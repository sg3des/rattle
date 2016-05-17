package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sg3des/wsrpc"
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

	ws := wsrpc.NewControllers(map[string]interface{}{"test": &TestController{}})
	http.Handle("/ws", ws)

	return http.ListenAndServe("127.0.0.1:8080", nil)
}

type TestController struct {
	Name string
	Text string
}

func (c *TestController) Index(r *wsrpc.Message) *wsrpc.Message {
	fmt.Println(c)
	return &wsrpc.Message{To: wsrpc.NewRPCMethod("test", "Recieve"), Data: []byte(`{"key":"value"}`)}
}
