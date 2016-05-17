'use strict';

var conn = new wsrpc.NewConnection("ws://127.0.0.1:8080/ws", true);

var test = {
	Send: function (msg) {
		var data = {};
		data.text = document.getElementById("text").value;
		data.name = "val_name"

		conn.send("test.Send", "test.Index", data)
	},

	Recieve: function (msg) {
		console.log("recieve msg:", msg)
	}
}