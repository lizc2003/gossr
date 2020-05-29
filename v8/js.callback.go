package v8

const callbackJsContent = `
var V8WorkerCallback = function() {
	var callbackMap = new Map();

	var v8c = {
		addCallback: function (type, func) {
			callbackMap.set(type, func);
		},

		recvCallback: function (type, msg) {
			try {
				var func = callbackMap.get(type);
				if (func !== undefined) {
					func(msg);
				}
			} catch (e) {
				console.error(e);
			}
		}
	};
	return v8c
};

var _v8workerCallback = new V8WorkerCallback();
v8worker.setRecv(_v8workerCallback.recvCallback);
`
