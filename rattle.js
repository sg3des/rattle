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
			return context[func]
		}

		var request = function (to, data) {
			if (to == "" || to == undefined) {
				return
			}

			if ("#+.".indexOf(to[0]) != -1) {
				var element = getElement(to)
				if (!element) {
					console.error("target element not found: ", to)
					return
				}

				var setvalue = false
				if (element.tagName == "INPUT" || element.tagName == "TEXTAREA") {
					setvalue = true
				}

				switch (to[0]) {
				case "#" || ".":
					if (setvalue) element.value = data
					else
						element.innerHTML = data

					break;
				case "+":
					if (setvalue) element.value += data
					else
						element.innerHTML += data

					break;
				}
				return
			}

			try {
				data = JSON.parse(data)
				func(to)(data)
			} catch (e) {
				func(to)(data)
			}

		}

		var getElement = function (name) {
			if (name[0] == "#") {
				return document.getElementById(name.slice(1)) || document.querySelector(name.slice(1))
			}

			if (name[0] == "+") {
				return document.querySelector(name.slice(1))
			}

			return document.querySelector(name)
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
					to = splitted[0],
					data = splitted.slice(1, splitted.length).join(" ")

				if (this.debug) console.log("rattle: Get message " + to + " with data length:", data.length)

				if (this.callbacks["message"]) {
					this.callbacks["message"](to, data)
				} else {
					request(to, data)
				}
			},

			on: function (name, callback) {
				this.callbacks[name] = callback;
			},

			send: function (to, data) {
				if (to == undefined || to == "") {
					console.warn("rattle: field 'To'(target function) is not filled")
					return
				}

				if (data == undefined) {
					data = {}
				}

				if (this.debug) console.log("rattle: Send message to: " + to + " with data:", data)

				this.ws.send(to + " " + JSON.stringify(data) + "\n")
			}

		} // end NewConnection.prototype

		global.rattle = {
			NewConnection: NewConnection
		}

	} else {
		console.warn("rattle: WebSockets not supported!")
	}
})(this)