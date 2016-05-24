'use strict';

var conn = new rattle.NewConnection("ws://127.0.0.1:8080/ws", true); //addr, debug

conn.on("open", function (evt) {
	conn.send("Main.Timer");

	construct.Send();
})

//for this case, rattle correct fill field `From` 
var test = {
	Send: function (msg) {
		var data = {};
		data.text = document.getElementById("text").value;
		data.name = "val_name";

		conn.send("Main.Index", data);
	},

	Recieve: function (msg) {
		console.log("recieve msg:", msg);
		document.getElementById("msgs").innerHTML = JSON.stringify(msg);
	}
}

//WARNING! for constructors, rattle not correct determine  field `From`!!!
function exConstructor(name) {
	this.name = name;

	this.Send = function (msg) {
		conn.send("main.something"); //call backend procedure not case sensitive - rattle auto TITLize incoming request on backend side, for example this will be transform to "Main.Something"
	};

	this.Recieve = function (msg) {
		document.getElementById("somefield").innerHTML = "i`m function Recieve"
	};
}

var construct = new exConstructor("construct");