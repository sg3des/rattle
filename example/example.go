package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sg3des/rattle"
)

func main() {
	//set debug mode
	rattle.Debug = true
	log.SetFlags(log.Lshortfile)

	r := rattle.NewRattle()
	r.AddRoute("index", index)
	r.AddRoute("text", text)
	r.AddRoute("upload", upload)

	//bind event on Connect
	r.SetOnConnect(onConnect)
	r.SetOnDisconnect(onDisconnect)

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/rattle/", http.StripPrefix("/rattle/", http.FileServer(http.Dir("../"))))
	http.Handle("/ws", r.Handler())

	fmt.Println("web server listen on 127.0.0.1:8080")

	http.ListenAndServe("127.0.0.1:8080", nil)
}

func index(r *rattle.Request) {
	msg := r.NewMessage("=#description", []byte(`Rattle is tiny websocket double-sided RPC framework, designed for create dynamic web applications`))
	msg.Send()
}

func text(r *rattle.Request) {
	log.Printf("text: % 02x", r.Data)
	r.NewMessage("=#msgs", r.Data).Send()
}

func upload(r *rattle.Request) {
	log.Println("incoming file:", r.File.Name, "size:", r.File.Buffer.Len())

	ioutil.WriteFile(r.File.Name, r.File.Buffer.Bytes(), 0644)

	r.NewMessage("+#msgs", []byte(fmt.Sprintf("file `%s` uploaded", r.File.Name))).Send()
}

//onConnect handler function for event connection
func onConnect(r *rattle.Request) {
	log.Println("someone is connected")

	// //clock
	// for {
	// 	r.NewMessage("=#clock", []byte(time.Now().Format("15:04:05"))).Send()
	// 	time.Sleep(1 * time.Second)
	// }
}

//onDisconnect handler
func onDisconnect(r *rattle.Request) {
	log.Println("someone is disconnected")
}
