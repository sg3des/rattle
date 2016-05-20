'use strict';

(function (global) {
	if (global["WebSocket"]) {

		var func = function (functionName, msg) {
			var namespaces = functionName.split(".")
			var func = namespaces.pop()
			var context = window

			for (var i = 0; i < namespaces.length; i++) {
				if (context[namespaces[i]] == undefined) {
					console.warn("rpc function not found", functionName)
					return
				}
				context = context[namespaces[i]]
			}
			return context[func] //(msg)
		}

		var request = function (msg) {
			if (msg.To == "" || msg.To == undefined) {
				return
			}

			if ("#+.".indexOf(msg.To[0]) != -1) {
				var element = get(msg.To)
				if (!element) {
					console.error("target element not found: ", msg.To)
					return
				}

				var setvalue = false
				if (element.tagName == "INPUT" || element.tagName == "TEXTAREA") {
					setvalue = true
				}

				switch (msg.To[0]) {
				case "#" || ".":
					if (setvalue) element.value = msg.Data
					else
						element.innerHTML = msg.Data

					break;
				case "+":
					if (setvalue) element.value += msg.Data
					else
						element.innerHTML += msg.Data

					break;
				}
				return
			}

			msg.Data = JSON.parse(msg.Data)
			func(msg.To)(msg)
		}

		var get = function (name) {
			if (name[0] == "#") {
				return document.getElementById(name.slice(1)) || document.querySelector(name.slice(1))
			}

			if (name[0] == "+") {
				return document.querySelector(name.slice(1))
			}

			return document.querySelector(name)
		}

		var getCallerFunc = function () {
			try {
				throw Error('')
			} catch (err) {

				var rpcfrom = err.stack.split("\n")[2].split("@")[0]
				if (rpcfrom == "") {
					rpcfrom = "global"
				}
				return rpcfrom
			}
		}

		var NewConnection = function (addr, debug) {
			this.debug = debug
			this.ws = new WebSocket(addr)
			this.ws.onclose = this.onclose.bind(this)
			this.ws.onopen = this.onopen.bind(this)
			this.ws.onmessage = this.onmessage.bind(this)
			this.callbacks = {}
		}

		NewConnection.prototype = {
			constructor: NewConnection,

			onopen: function (evt) {
				if (this.debug) console.log("rattle: connected")
				if (this.callbacks["open"]) this.callbacks["open"](evt);
			},

			onclose: function (evt) {
				if (this.debug) console.log("rattle: disconnected")
				if (this.callbacks["close"]) this.callbacks["close"](evt);
			},

			onmessage: function (incomingData) {
				var splitted = incomingData.data.split(' '),
					from = splitted[0],
					to = splitted[1],
					data = splitted.slice(2, splitted.length).join(" ")

				if (this.debug) console.log("rattle: Get message " + from + "->" + to + " with data length:", data.length)

				var msg = {
					From: from,
					To: to,
					Data: data
				}

				if (this.callbacks["message"]) {
					this.callbacks["message"](msg, evt)
				} else {
					request(msg)
					// execute(to, msg)
				}
			},

			on: function (name, callback) {
				this.callbacks[name] = callback;
			},

			send: function (msg) {
				if (msg["To"] == undefined || msg["To"] == "") {
					console.warn("rattle: field 'To'(target function) is not filled")
					return
				}

				if (msg["From"] == undefined) {
					msg["From"] == getCallerFunc();
				}

				if (msg["Data"] == undefined) {
					msg["Data"] = {}
				}

				if (this.debug) console.log("rattle: Send message " + msg.From + "->" + msg.To + " with data:", msg.Data)

				this.ws.send(msg.From + " " + msg.To + " " + JSON.stringify(msg.Data) + "\n")
			}

		} // end NewConnection.prototype

		global.rattle = {
			NewConnection: NewConnection
		}

	} else {
		console.warn("rattle: WebSockets not supported!")
	}
})(this)