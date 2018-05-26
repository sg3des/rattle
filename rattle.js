'use strict';

(function (global) {
	if (global["WebSocket"]) {

		var call = function (to, data) {
			if (to == "" || to == undefined) return;

			if ("=+@".indexOf(to[0]) != -1) {
				callElement(to, data)
			} else {
				callFunction(to, data)
			}
		}

		var callElement = function (to, data) {
			var element
			switch (to[1]) {
			case "#":
				element = document.getElementById(to.slice(2))
				break;
			default:
				element = document.querySelector(to.slice(1))
				break;
			}

			if (!element) {
				console.error("target element not found: ", to.slice(1))
				return;
			}

			if (to[0] == "@") {
				element.outerHTML = data
			} else {
				if (element.tagName == "INPUT" || element.tagName == "TEXTAREA") {
					element.value = to[0] == "=" ? data : element.value + data
				} else {
					element.innerHTML = to[0] == "=" ? data : element.innerHTML + data
				}
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

		var NewConnection = function (addr, debug) {
			this.addr = addr
			this.debug = debug
			this.connected = false
			this.boundary = ""

			this.ws = new WebSocket(addr)
			this.ws.onopen = this.onConnect.bind(this)
			this.ws.onclose = this.onDisconnect.bind(this)
			this.ws.onmessage = this.onMessage.bind(this)

			this.callbacks = {}

			this.connect = connect
			this.disconnect = disconnect

			this.stream = stream
			this.streamNext = streamNext
		}

		NewConnection.prototype = {
			constructor: NewConnection,

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
					to = splitted[0].trim(),
					data = splitted.slice(1, splitted.length).join(" ").trim()

				if (to == "stream") {
					this.stream()
					return
				}

				if (this.debug) console.log("rattle: Get message to: " + to + " with data length:", data.length)

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

				var msg = {}
				msg.to = to
				msg.url = window.location.href
				

				if (data != undefined) {
					msg.data = data
					msg.type = typeof data
				}

				//reconnect if current not connected
				if (this.ws.readyState == 3) {
					this.connect()
				}

				this.ws.send(JSON.stringify(msg) + "\n")

				if (this.debug) console.log("rattle: Send message to: " + to + " with data:", JSON.stringify(data))
			},

			file: function (to, input, userdata) {
				if (input.files.length == 0) {
					console.warn("files not found")
					return
				}

				streamData.to = to
				streamData.url = window.location.href
				streamData.files = input.files
				streamData.userdata = userdata
				streamData.current = 0
				streamData.offset = 0
				streamData.i = 0

				if (this.debug) console.log("rattle: open stream", streamData)

				this.stream()
			},
		} // end newConnection.prototype

		var streamData = {
			to: "",
			url: "",
			files: "",
			userdata: {},
			i: 0,
			offset: 0,
			slicesize: 1024 * 1024
		}

		var streamNext = function (ws) {
			if (streamData.i >= streamData.files.length) {
				return false;
			}
			// console.log(this)
			switch (streamData.offset) {
			case 0:
				// console.log(this)
				var msg = {
					to: streamData.to,
					url: window.location.href,
					type: "stream",
					json: streamData.userdata,
					stream: {
						name: streamData.files[streamData.i].name,
						size: streamData.files[streamData.i].size,
						slicesize: streamData.slicesize
					}
				}

				ws.send(JSON.stringify(msg) + "\n")

				// ws.send(streamData.to + " stream " + JSON.stringify(streamData.userdata) + " " + JSON.stringify({
				// 	name: streamData.files[streamData.i].name,
				// 	size: streamData.files[streamData.i].size,
				// 	slicesize: streamData.slicesize
				// }) + "\n")

				return true;

			case streamData.files[streamData.i].size:
				ws.send("\n{\"type\":\"finish\"}\n")

				streamData.offset = 0
				streamData.i++

				return streamNext(ws);
			default:
				ws.send("\n{\"type\":\"chunk\"}\n")
				return true;
			}
		}

		var stream = function () {
			if (!streamNext(this.ws)) {
				return
			}

			if (this.debug) console.log("rattle: send chunk ", streamData)

			var offsetEnd = streamData.offset + streamData.slicesize
			if (offsetEnd > streamData.files[streamData.i].size) {
				offsetEnd = streamData.files[streamData.i].size
			}

			//reconnect if current not connected
			if (this.ws.readyState == 3) {
				this.connect()
			}

			this.ws.send(streamData.files[streamData.i].slice(streamData.offset, offsetEnd))
			streamData.offset = offsetEnd
		}

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
			NewConnection: NewConnection
		}

	} else {
		console.warn("rattle: WebSockets not supported!")
	}
})(this)