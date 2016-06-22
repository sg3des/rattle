'use strict';

var conn = new rattle.NewConnection("ws://127.0.0.1:8080/ws", true); //addr, debug

console.log(conn)
conn.event("onConnect", function (evt) {
	document.getElementById("toggle").checked = true;
	conn.send("Main.Timer");
	conn.send("Main.Index")
})

var test = {
	Send: function () {
		var data = {};
		data.text = document.getElementById("text").value;

		var url = document.getElementById("json").checked ? "Main.JSON" : "Main.RAW"

		conn.send(url, data);
	},

	RecieveJSON: function (data) {
		document.getElementById("msgs").innerHTML = JSON.stringify(data);
	},

	RecieveRAW: function (data) {
		document.getElementById("msgs").innerHTML = data;
	}
}

function toggle() {
	if (conn.connected) {
		conn.disconnect()
	} else {
		conn.connect()
	}
}