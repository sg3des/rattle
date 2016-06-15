[![Build Status](https://travis-ci.org/sg3des/rattle.svg?branch=master)](https://travis-ci.org/sg3des/rattle)

# RATTLE

### Rattle is websocket double-sided rpc solution - designed for create MVVM web applications.

#### WARNING: Rattle is only concept! 

## INSTALL

	go get github.com/sg3des/rattle

run example:

	cd $GOPATH/src/github.com/sg3des/rattle/example/ 
	go run example.go

web server will be listen at `127.0.0.1:8080`


## IDEA

MVC architecture with http requests not always suitable for web applications! 
MVVM architecture more suitable for this, but design it on http requests a bit embarrassing, then need use websockets! 

Rattle is tiny websocket double-sided framework. Backend is go, Frontend - javascript.

	1) For connection use only websocket;
	2) Requests can come from backend and frontend at the same time;
	3) Requests are processed asynchronously;
	4) Supports file uploading, multiple and large files too;
	5) RATTLE significantly reduces the amount of code in web application;


## USAGE

#### Backend:

First need set controllers, and add http requests handler:

```
	wshandle := rattle.SetControllers(&Main{})
	http.Handle("/ws", wshandle)
```
Where: 
* `&Main{}` is *link* on your controller;
* in real porject, recommended located controllers in another package, then *link* may be the: `&controllers.Main{}`;
* may to specify many controllers: `rattle.SetControllers(&Main{},&Second{},&Third{})`.


Write controller:

```
	type Main struct {
		Name string
		Text string
	}
```
Controllers is just struct where parsed JSON requests, however this it can be left empty: `type Main struct{}`
* Controllers always must be a public,  i.e begin with a capital letter.

Write method for controller:
```
func (c *Main) Index(r *rattle.Conn) *rattle.Message {
	r.NewMessage("#h1", []byte(`Main Index`)).Send()
	return r.NewMessage("test.Recieve", []byte(`{"newJSONkey":"`+c.Text+`"}`))
}

```
* Methods always takes incoming messages, and can(not necessary) return response;
* based on an incoming message, you can create an answer: `r.NewMessage(to,data)`, where:
	* first argument is can the name of the called frontend function, or if it starts with symbols `=` or `+` or - target will be HTML element found with js `querySelector`, examples:
		* `=#idname`, `=tagname`, `=.classname` - crop first symbol `=`, then **place** data to element founded by `querySelector`
		* `+#idname`, `+tagname`, `+.classname` - crop first symbol `+`, then **adds** the data to the existing in founded element;
	* second argument is data in []byte format, and may be type JSON, HTML, etc...


Strucutre of message:
```
type Message struct {
	[not exported field with currect Connection]
	To   []byte
	Data []byte
}
```
* **To** field contain name of the called function - is required to fill by user;
* **Data** field: 
	* for messages from **backend** to **frontend** can contain any type of data: HTML, JSON, etc;
	* for messages from **frontend** to **backend** always JSON.


#### Frontend:
First need connect to server/backend:

```
	var conn = new rattle.NewConnection("ws://127.0.0.1:8080/ws", true);
```
* second boolean argument is enable/disable debug mode.

Possible bind some custom actions for events: **onConnect**,**onDisconnect**,**onMessage**. In the next example bind event *onConnect*. 
```
	conn.event("onConnect", function (evt) {
		conn.send("Main.Index");
	})

```

In order to send request/message:
```
	conn.send("Main.Index", data);
```

Write frontend controllers:
```
var test = {
	var Recieve = function(msg) {
		[...]
	}
}
```

or use constructors:
```
function constructorExample(){
	this.Recieve = function(msg) {
		[...]
	}
}

var test = new constructorExample()
```

### BENCHMARKS

	BenchmarkJSONRequests-8 	  200000	      7329 ns/op
	BenchmarkEmptyRequests-8	  300000	      4196 ns/op




## TODO

* need more test for backend;
* need tests for frontend;
* and many more other - this is yet a concept!
