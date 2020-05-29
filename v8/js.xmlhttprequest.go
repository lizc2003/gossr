package v8

const xmlHttpRequestJsContent = `
var V8HttpRequestHandler = function() {
	var httpMap = new Map();

	httpCallback = function(obj, msg) {
		var evt = msg.event;
		if (evt === "onfinish") {
			httpMap.delete(msg.httpid);
		} else if (evt === "onerror") {
			obj._onErrorCallback(msg.error)
		} else if (evt === "onstart") {
			obj._onStartCallback()
		} else if (evt === "onheader") {
			obj._onHeaderCallback(msg.status, msg.headers)
		} else if (evt === "onend") {
			obj._onEndCallback(msg.response)
		}
	};

	var v8c = {
		addHttpObj: function (httpid, obj) {
			httpMap.set(httpid, obj);
		},

		recvCallback: function (_msg) {
			var msg = JSON.parse(_msg);
			var obj = httpMap.get(msg.httpid);
			if (obj !== undefined) {
				httpCallback(obj, msg);
			}
		}
	};
	return v8c
};

var _v8HttpRequestHandler = new V8HttpRequestHandler()
_v8workerCallback.addCallback(20, _v8HttpRequestHandler.recvCallback)

var XMLHttpRequest = function() {
	var method, url,
		xhr, callEventListeners,
		statusCodes,
		httpid = 0,
		listeners = ['readystatechange', 'abort', 'error', 'loadend', 'progress', 'load'],
		privateListeners = {},
		responseHeaders = {},
		headers = {};

	statusCodes = {
		100:'Continue',101:'Switching Protocols',102:'Processing',200:'OK',201:'Created',202:'Accepted',203:'Non-Authoritative Information',204:'No Content',205:'Reset Content',206:'Partial Content',207:'Multi-Status',208:'Already Reported',226:'IM Used',300:'Multiple Choices',301:'Moved Permanently',302:'Found',303:'See Other',304:'Not Modified',305:'Use Proxy',306:'Switch Proxy',307:'Temporary Redirect',308:'Permanent Redirect',400:'Bad Request',401:'Unauthorized',402:'Payment Required',403:'Forbidden',404:'Not Found',405:'Method Not Allowed',406:'Not Acceptable',407:'Proxy Authentication Required',408:'Request Timeout',409:'Conflict',410:'Gone',411:'Length Required',412:'Precondition Failed',413:'Request Entity Too Large',414:'Request-URI Too Long',415:'Unsupported Media Type',416:'Requested Range Not Satisfiable',417:'Expectation Failed',418:'I\'m a teapot',419:'Authentication Timeout',420:'Method Failure',420:'Enhance Your Calm',422:'Unprocessable Entity',423:'Locked',424:'Failed Dependency',426:'Upgrade Required',428:'Precondition Required',429:'Too Many Requests',431:'Request Header Fields Too Large',440:'Login Timeout',444:'No Response',449:'Retry With',450:'Blocked by Windows Parental Controls',451:'Unavailable For Legal Reasons',451:'Redirect',494:'Request Header Too Large',495:'Cert Error',496:'No Cert',497:'HTTP to HTTPS',498:'Token expired/invalid',499:'Client Closed Request',499:'Token required',500:'Internal Server Error',501:'Not Implemented',502:'Bad Gateway',503:'Service Unavailable',504:'Gateway Timeout',505:'HTTP Version Not Supported',506:'Variant Also Negotiates',507:'Insufficient Storage',508:'Loop Detected',509:'Bandwidth Limit Exceeded',510:'Not Extended',511:'Network Authentication Required',520:'Origin Error',521:'Web server is down',522:'Connection timed out',523:'Proxy Declined Request',524:'A timeout occurred',598:'Network read timeout error',599:'Network connect timeout error'
	};

	callEventListeners = function(listeners) {
		var listenerFound = false, i;
		if (typeof listeners === 'string') {
			listeners = [listeners];
		}
		listeners.forEach(function(e) {
			if (typeof xhr['on' + e] === 'function') {
				listenerFound = true;
				xhr['on' + e].call(xhr);
			}
			if (privateListeners.hasOwnProperty(e)) {
				for (i = 0; i < privateListeners[e].length; i++) {
					if (privateListeners[e][i] !== void 0) {
						listenerFound = true;
						privateListeners[e][i].call(xhr);
					}
				}
			}
		});
		return listenerFound;
	};

	xhr = {
		UNSENT:		   0,
		OPENED:		   1,
		HEADERS_RECEIVED: 2,
		LOADING:		  3,
		DONE:			 4,

		status: 0,
		statusText: '',
		responseText: '',
		response: '',
		readyState: 0,
		timeout: 2000,

		open: function(_method, _url) {
			if (_method === void 0 || _url === void 0) {
				throw TypeError('Failed to execute \'open\' on \'XMLHttpRequest\': 2 arguments required, but only ' + +(_method || url) + ' present.');
			}
			method = _method;
			url = _url;
			xhr.readyState = xhr.OPENED;
		},

		send: function(post) {
			if (xhr.readyState !== xhr.OPENED) {
				throw {
					name: 'InvalidStateError',
					message: 'Failed to execute \'send\' on \'XMLHttpRequest\': The object\'s state must be OPENED.'
				}
			}

			var options = {
				cmd: "open",
				url: url,
				method: method,
				headers: headers
			};
 
			if (post) {
				var ptype = typeof(post);
				if (ptype == 'object') {
					try {
						var data = ""
						for (var key in post) {
							if (Object.prototype.hasOwnProperty.call(post, key)) {
								if (data.length > 0) {
									data += "&";
								}
								data += key.toString() + '=' + encodeURIComponent(post[key]);
							}
						}
						post = data;
					} catch(e) {
						post = "";
					}
				} else if (ptype !== "string") {
					post = "";
				}
				if (post.length > 0) {
					//console.log("post data:" + post)
					options.headers['content-length'] = post.length.toString();
					options.post = post
				}
			}
 
			httpid = parseInt(v8worker.request(11, JSON.stringify(options)));
			if (httpid > 0) {
				_v8HttpRequestHandler.addHttpObj(httpid, xhr)
			} else {
				throw TypeError('Failed to execute \'send\' on \'XMLHttpRequest\'');
			}
		},

		setRequestHeader: function(header, value) {
			if (header === void 0 || value === void 0) {
				throw TypeError(' Failed to execute \'setRequestHeader\' on \'XMLHttpRequest\': 2 arguments required, but only ' + +(headers || value) + ' present.');
			}
			headers[header] = value;
		},

		abort: function() {
			if (httpid > 0) {
				callEventListeners('abort');

				var options = {
					cmd: "abort",
					httpid: httpid
				};
				v8worker.request(11, JSON.stringify(options));
			}
		},

		addEventListener: function(type, fn) {
			if (listeners.indexOf(type) !== -1) {
				privateListeners[type].push(fn);
			}
		},

		removeEventListener: function(type, fn) {
			var index;
			if (privateListeners[type] === void 0) {
				return;
			}
			while ((index = privateListeners[type].indexOf(fn)) !== -1) {
				privateListeners[type].splice(index, 1);
			}
		},

		getAllResponseHeaders: function() {
			var res = '';
			for (var key in responseHeaders) {
				res += key + ': ' + responseHeaders[key] + '\n';
			}
			return res.slice(0, -1);
		},

		getResponseHeader: function(key) {
			if (responseHeaders.hasOwnProperty(key)) {
				return responseHeaders[key];
			}
			return null;
		},

		_onStartCallback: function() {
			callEventListeners('loadstart');
		},

		_onHeaderCallback: function(status, headers) {
			xhr.status = status;
			if (statusCodes.hasOwnProperty(xhr.status)) {
				xhr.statusText = statusCodes[xhr.status];
			} else {
				xhr.statusText = xhr.status.toString();
			}
			responseHeaders = headers;
			xhr.readyState = xhr.HEADERS_RECEIVED;
			callEventListeners('readystatechange');
			xhr.readyState = xhr.LOADING;
			callEventListeners('readystatechange', 'progress');
		},

		_onErrorCallback: function(err) {
			if (callEventListeners('error') === false) {
				console.error(err);
			}
			callEventListeners('loadend');
		},

		_onEndCallback: function(response) {
			xhr.readyState = xhr.DONE;
			xhr.responseText = response;
			xhr.response = response;
			callEventListeners(['readystatechange', 'load', 'loadend']);
		}
	};

	for (var i = 0; i < listeners.length; i++) {
		xhr['on' + listeners[i]] = null;
		privateListeners[listeners[i]] = [];
	}

	return xhr;
};
`
