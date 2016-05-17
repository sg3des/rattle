'use strict';

(function (global) {
	if (global["WebSocket"]) {

		var execute = function (functionName, context /*, args */ ) {
			var args = Array.prototype.slice.call(arguments, 2);
			var namespaces = functionName.split(".");
			var func = namespaces.pop();
			for (var i = 0; i < namespaces.length; i++) {
				context = context[namespaces[i]];
			}
			return context[func].apply(context, args);
		}

		var NewConnection = function (addr, debug) {

			this.ws = new WebSocket(addr);

			this.debug = debug;

			this.ws.onclose = this.onclose.bind(this);
			this.ws.onopen = this.onopen.bind(this);
			this.ws.onmessage = this.onmessage.bind(this);
		}

		NewConnection.prototype = {
			constructor: NewConnection,

			onopen: function (evt) {
				if (this.debug) {
					console.log("wsrpc: connected");
				}
			},

			onclose: function (evt) {
				if (this.debug) {
					console.log("wsrpc: disconnected");
				}
			},

			onmessage: function (incomingData) {
				var splitted = incomingData.data.split(' '),
					from = splitted[0],
					to = splitted[1],
					data = splitted.slice(2, splitted.length).join(" ");

				if (this.debug) {
					console.log("wsrpc: Get message " + from + "->" + to + " with data:", data);
				}

				var msg = {
					From: from,
					To: to,
					Data: JSON.parse(data)
				};

				execute(to, window, msg);
			},

			send: function (from, to, data) {
				if (this.debug) {
					console.log("wsrpc: Send message " + from + "->" + to + " with data:", data)
				}

				this.ws.send(from + " " + to + " " + JSON.stringify(data) + "\n");
			}

		} // end NewConnection.prototype

		global.wsrpc = {
			NewConnection: NewConnection
		};

	} else {
		console.warn("wsrpc: WebSockets not supported!");
	}
})(this)