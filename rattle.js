'use strict';

(function (global) {
	if (global["WebSocket"]) {

		var call = function (to, data) {
			if (to == "" || to == undefined) return;

			if ("=+".indexOf(to[0]) != -1) {
				callElement(to, data)
			} else {
				callFunction(to, data)
			}
		}

		var callElement = function (to, data) {
			var element = document.querySelector(to.slice(1))
			if (!element) {
				console.error("target element not found: ", to.slice(1))
				return;
			}

			if (element.tagName == "INPUT" || element.tagName == "TEXTAREA") {
				element.value = to[0] == "=" ? data : element.value + data
			} else {
				element.innerHTML = to[0] == "=" ? data : element.innerHTML + data
			}
		}

		var callFunction = function (to, data) {
			var func = getFunction(to)
			if (!func) {
				console.error("target function not found: ", to)
				return;
			}

			try {
				data = JSON.parse(data)
				func(data)
			} catch (e) {
				func(data)
			}
		}

		var getFunction = function (functionName) {
			var namespaces = functionName.split(".")
			var func = namespaces.pop()
			var context = window

			for (var i = 0; i < namespaces.length; i++) {
				if (context[namespaces[i]] == undefined) {
					// console.warn("rpc function not found", functionName)
					return undefined;
				}
				context = context[namespaces[i]]
			}
			return context[func];
		}

		var newConnection = function (addr, debug) {
			this.addr = addr
			this.debug = debug
			this.connected = false

			this.ws = new WebSocket(addr)
			this.ws.onopen = this.onConnect.bind(this)
			this.ws.onclose = this.onDisconnect.bind(this)
			this.ws.onmessage = this.onMessage.bind(this)

			this.callbacks = {}

			this.connect = connect
			this.disconnect = disconnect

		}

		newConnection.prototype = {
			constructor: newConnection,

			onConnect: function (evt) {
				this.connected = true
				if (this.debug) console.log("rattle: connected")
				if (this.callbacks["onConnect"]) this.callbacks["onConnect"](evt);
			},

			onDisconnect: function (evt) {
				this.connected = false
				if (this.debug) console.log("rattle: disconnected")
				if (this.callbacks["onDisconnect"]) this.callbacks["onDisconnect"](evt);
			},

			onMessage: function (incomingData) {
				var splitted = incomingData.data.split(' '),
					to = splitted[0],
					data = splitted.slice(1, splitted.length).join(" ")

				if (this.debug) console.log("rattle: Get message " + to + " with data length:", data.length)

				if (this.callbacks["onMessage"]) {
					this.callbacks["onMessage"](to, data)
				} else {
					call(to, data)
				}
			},

			event: function (name, callback) {
				this.callbacks[name] = callback;
			},

			send: function (to, data) {
				if (to == undefined || to == "") {
					console.warn("rattle: field 'To'(target function) is not filled")
					return;
				}

				if (data == undefined) {
					data = {}
				}

				if (this.debug) console.log("rattle: Send message to: " + to + " with data:", data)

				this.ws.send(to + " " + JSON.stringify(data) + "\n")
			}
		} // end newConnection.prototype

		var disconnect = function () {
			if (this.connected) {
				this.ws.close()
			}
		}

		var connect = function () {
			if (this.connected) {
				return
			}
			this.ws = new WebSocket(this.addr)
			this.ws.onopen = this.onConnect.bind(this)
			this.ws.onclose = this.onDisconnect.bind(this)
			this.ws.onmessage = this.onMessage.bind(this)
		}

		global.rattle = {
			newConnection: newConnection
		}

	} else {
		console.warn("rattle: WebSockets not supported!")
	}
})(this)