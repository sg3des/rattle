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

* For connections use websockets;
* Requests can come from backend and frontend at the same time: 
	* From Backend possible call any Frontend(js) functions and pass arguments.
	* From Backend possible directly insert data to any html element.
	* From Frontend possible call any public methods of declare controllers.


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
* Controllers always must be a public, i.e begin with a capital letter.

Write method for controller:
```
func (c *Main) Index(r *rattle.Message) *rattle.Message {
	r.NewMessage("#h1", []byte(`Main Index`)).Send()
	return r.NewMessage("test.Recieve", []byte(`{"newJSONkey":"`+c.Text+`"}`))
}

```
* Methods always takes incoming messages, and can(not necessary) return response;
* based on an incoming message, you can create an answer: `r.NewMessage(to,data)`, where:
	* first argument is can the name of the called frontend function, or if it starts with symbols `#`, `+` or `.` - target HTML element (**!NOTICE: symbol selector is not fully CSS compatible!**), that means: 
		* `#` - crop first symbol, then search element by id, or if it not found search with querySelector, after this place data into element;
		* `+` - crop first symbol, then search with query selector, and then adds the html data to the existing in element;
		* `.` - just search with query selector, and then place the data into this element.
	* second argument is data in []byte format, and may be type JSON, HTML, etc...


Strucutre of message:
```
type Message struct {
	WS   *websocket.Conn
	From []byte
	To   []byte
	Data []byte
}
```
* **WS** field contains current websocket connection - does not require user action;
* **From** field is name of function which send request - rattle autofill this field;
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

Possible bind some custom actions for events, in the next example bind event *open*. 
```
	conn.on("open", function (evt) {
		[...]
	})

```

In order to send request/message:
```
	conn.send({
		To: "Main.Index", // name of the called backend function
		Data: data //some data, not necessary
	});
```
* !WARNING: Field `From` is likely for reference, Rattle autofill field `From`, but for constructors this will be incorrect, in this case you can manually fill this field, or just ignore it.

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
* !WARNING: keep in mind that in this case rattle is not correct fill field `From`!


## TODO:

* need tests for frontend;
* and many more other - this is yet a concept!
