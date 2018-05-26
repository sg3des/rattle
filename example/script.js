'use strict';

var r = new rattle.NewConnection("ws://127.0.0.1:8080/ws", true); //addr, debug


r.event("onConnect", function (evt) {
	r.send("index");
})

function send(str) {
	r.send("text", str);
}

function upload(input) {
	r.file("upload", input, {"key":"data"});
}