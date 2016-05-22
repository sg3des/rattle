# RATTLE

### Rattle is websocket double-sided rpc solution - designed for create MVVM web applications.

#### WARNING: Rattle is only concept!

## INSTALL

	go get github.com/sg3des/rattle

run example:

	go run $GOPATH/src/github.com/sg3des/rattle/example/example.go

web server will be listen at `127.0.0.1:8080`


## IDEA

MVC architecture with http requests not always suitable for web applications! 
MVVM architecture more suitable for this, but design it on http requests a bit embarrassing, then need use websockets! 

Rattle is tiny websocket double-sided framework. Backend is go, Frontend - javascript.

* For connections use websockets;
* Requests can come from backend and frontend at the same time: 
	* From Backend possible call any Frontend(js) functions and pass arguments.
	* From Backend possible directly insert data to any html element.
	* From Frontend possible call any public methods of declare controllers.


## USAGE

#### Backend:

First need set controllers:

```
	//bind controllers and get handler
	wshandle := rattle.SetControllers(&Main{})
	http.Handle("/ws", wshandle)

	//Main controller, into fields parse JSON requests
	type Main struct {
		Name string
		Text string
	}

	//Index is method of controller Main on request takes incoming message and possible return answer message
	func (c *Main) Index(r *rattle.Message) *rattle.Message {

		//just send message - insert data to h1 field
		r.NewMessage("#h1", []byte(`Main Index`)).Send()

		//returned message call test.Recieve Frontend function with JSON data
		return r.NewMessage("test.Recieve", []byte(`{"newJSONkey":"`+c.Text+`"}`))
	}

```

Controller is a struct, in which the fields will be parsed queries.
Methods takes incoming messages, and may(!not necessary) return messages for answer.

Message structure contains current connection in **WS** field, **From** field is name of function which send request, **To** field contain name of the called function, field **Data**: for backend is must be JSON for frontend is may be custom data

```
	type Message struct {
		WS   *websocket.Conn
		From []byte
		To   []byte
		Data []byte
	}
```



#### Frontend:
First need connect to server/backend

```
	var conn = new rattle.NewConnection("ws://127.0.0.1:8080/ws", true);
```

Possible bind some custom actions for events, in next example bind event *open*. 
```
	conn.on("open", function (evt) {
		[***]
	})

```

In order to send request/message:
```
	conn.send({
		To: "Main.Index", // name of the called backend function
		Data: data //some data, not necessary
	});
```

!NOTICE: Field "From" is likely for reference, Rattle autofill field "From", but for constructors this will be incorrect, in this case you can manually fill this field, or just ignore it.
