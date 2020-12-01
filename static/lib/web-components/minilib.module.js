// bluemir's light-weight web component library.
// TODO import * as $ from "minilib"

var $ = {
	get: function(target, query) {
		if(typeof target.querySelector !== "function") {return $.get(document, target)}
		return target.querySelector(query);
	},
	all: function(target, query) {
		if(typeof target.querySelectorAll !== "function") {return $.all(document, target)}
		return Array.prototype.slice.call(target.querySelectorAll(query));
	},
	create: function(tagname, attr) {
		var newTag = document.createElement(tagname);
		if (attr && attr.$text){
			newTag.appendChild(document.createTextNode(attr.$text));
		}
		if (attr && attr.$html){
			newTag.innerHTML = attr.$html;
		}
		if (attr && attr.$child) {
			newTag.appendChild(attr.$child)
		}
		for(var key in (attr || {})){
			if (key[0] == "$") {
				continue; //skip
			}
			newTag.setAttribute(key, attr[key]);
		}
		return newTag;
	},
	request: async function $request(method, url, options) {
		var o = options || {}
		try {
			var opts = $.config.hook.preRequest(method, url, o) || o;
		} catch(e) {
			var opts = o;
		}

		if (opts.timestamp !== false) {
			opts.query = opts.query || {};
			opts.query["_timestamp"] = Date.now();
		}

		return new Promise(function(resolve, reject) {
			var req = new XMLHttpRequest();

			req.addEventListener("readystatechange", function(){
				if (req.readyState  == 4) {
					var result = {
						statusCode: req.status,
						text : req.responseText,
					};

					var contentType = req.getResponseHeader("Content-Type") || "";
					if(contentType.includes("application/json")) {
						result.json = JSON.parse(result.text);
					}

					if (req.status >= 200, req.status < 300){
						resolve(result)
					} else {
						reject(result);
					}
				}
			});

			if (opts.auth) {
				console.debug("request with auth", opts.auth)
				// In Chrome and firefox Auth header not included request(due to security, see https://bugs.chromium.org/p/chromium/issues/detail?id=128323)
				// so forced set header
				req.open(method, resolveParam(url, opts.params) + queryString(opts.query), true, opts.auth.user, opts.auth.password);
				req.setRequestHeader("Authorization", "Basic " + btoa(opts.auth.user+":"+opts.auth.password));
			} else {
				req.open(method, resolveParam(url, opts.params) + queryString(opts.query), true);
			}

			Object.keys(opts.header || {}).forEach(function(name){
				req.setRequestHeader(name, opts.header[name]);
			});

			switch (typeof opts.body) {
				case "object":
					req.setRequestHeader("Content-Type", "application/json")
					req.send(JSON.stringify(opts.body))
					break;
				case "string":
					req.send(opts.body);
					break;
				case "undefined":
					req.send();
					break; // just skip
				default:
					reject("unknown type: req.body");
					break;
			}
		});
	},
	timeout: async function(ms) {
		return new Promise(function(resolve, reject){
			setTimeout(resolve, ms);
		});
	},
	defer: function() {
		var ret = {}
		ret.promise = new Promise(function(resolve, reject){
			ret.resolve = resolve;
			ret.reject = reject;
		});
		return ret;
	},
	prevent: function(func){
		return function(evt){
			evt.preventDefault();
			return func();
		}
	},
	form: function(form) {
		var fd = new FormData(form)
		return Array.from(fd).reduce((obj, [k, v] )=> {
			switch($.get(form, `[name=${k}]`).attr("type")) {
				case "number":
					obj[k] = v-0;
					break;
				default:
					obj[k] = v;
					break;
			}
			return obj;
		}, {});
	},
	animateFrame: function(callback, {fps = 30} = {}) {
		var stop = false;
		var fpsInterval = 1000 / fps;
		var then = Date.now();
		animate();

		function animate() {
			if (stop) {
				return;
			}
			requestAnimationFrame(animate);

			var now = Date.now();
			var elapsed = now - then;

			if (elapsed > fpsInterval) {
				then = now - (elapsed % fpsInterval);

				var ret = callback(elapsed - (elapsed%fpsInterval));
				if (ret && ret.stop) {
					 stop = true;
				}
			}
		}
	},
	query: function(data, query) {
		var keys = query.split("\\.").map(str => str.split(".")).reduce((p, c) => {
			if (p.length == 0 ) {
				return c;
			}
			var last = p.pop();
			var first = c.shift();

			return [].concat(p, [last+"."+first], c);
		});

		if (query[0] == ".") {
			keys.shift() // remove first empty key
		}

		var result = data;
		for (var i = 0; i < keys.length; i++) {
			if (result === undefined) {
				return undefined
			}
			result = result[keys[i]];
		}
		return result;
	},
	querySet: function(obj, query, value) {
		var keys = query.split("\\.").map(str => str.split(".")).reduce((p, c) => {
			if (p.length == 0 ) {
				return c;
			}
			var last = p.pop();
			var first = c.shift();

			return [].concat(p, [last+"."+first], c);
		});

		if (query[0] == ".") {
			keys.shift() // remove first empty key
		}

		var visitor = obj;
		for (var i = 0; i < keys.length-1; i++) {
			if (visitor === undefined) {
				return undefined
			}
			visitor = visitor[keys[i]];
		}
		visitor[keys[keys.length - 1]] = value;
	},
	util: {
		wsURL: function(url){
			var u= new URL(url, document.location)
			u.protocol = document.location.protocol.includes("https") ? "wss:" : "ws:"
			return u;
		},
		filter: {
			notNull: e => e != null,
		},
		reduce: {
			appendChild: function(parent, child) {
				parent.appendChild(child);
				return parent;
			},
		},
	},
	event: new EventTarget(),
	_registerGlobal: function() {
		window.$ = this;
	},
	config: {
		hook: {
			preRequest: function(method, url, opt) { return opt }
		},
	},
}

function resolveParam(url, params) {
	if (params == null) {
		return url
	}
	return url.replace(/:([a-zA-Z0-9]+)/g, function(matched, name){
		if (params[name]) {
			return params[name];
		}
		console.warn(`[$.reqeust] find param pattern '${name}', but not provided`);
		return matched;
	});
}

function queryString(obj) {
	if (obj == null) {
		return "";
	}
	return "?" + Object.keys(obj).map(function(key) {
		return key + "=" + obj[key];
	}).join("&");
}

Object.keyValues= function(obj, f) {
	return Object.entries(obj).map(([key, value]) => {
		return {key, value};
	});
}

function extend(TargetClass, proto){
	if (TargetClass.hasOwnProperty("__minilib_inserted__")) {
		console.trace("already installed")
		return // already inserted
	}

	Object.keys(proto).forEach(function(name) {
		if (name  in TargetClass.prototype) {
			console.warn(`cannot extend prototype: '${name}' already exist`)
			return; // skip
		}
		TargetClass.prototype[name] = proto[name];
	});

	TargetClass.__minilib_inserted__ = true
}
extend(Node, {
	appendTo: function(target) {
		target.appendChild(this);
	},
	clear : function(filter) {
		var f = filter || function(e) { return true };
		this.childNodes.filter(f).forEach((e) => this.removeChild(e))
		return this;
	},
});
extend(Element, {
	attr: function(name, value){
		if (value === null) {
			this.removeAttribute(name);
			return
		}
		if (value !== undefined) {

			this.setAttribute(name, value)
			return value;
		} else {
			return this.getAttribute(name)
		}
	},
})

extend(EventTarget, {
	on: function(name, handler, opt) {
		this.addEventListener(name, handler, opt);

		return this;
	},
	off: function(name, handler, opt) {
		this.removeEventListener(name, handler, opt)

		return this;
	},
	fireEvent: function(name, detail) {
		var evt = new CustomEvent(name, {detail: detail});
		this.dispatchEvent(evt);
		return this;
	}
});

extend(NodeList, {
	"map":    Array.prototype.map,
	"filter": Array.prototype.filter,
	//"forEach": Array.prototype.forEach,
});
extend(HTMLCollection, {
	"map":     Array.prototype.map,
	"filter":  Array.prototype.filter,
	"forEach": Array.prototype.forEach,
});

extend(Array, {
	"promise": function() {
		var arr = this;
		return {
			all:  () => Promise.all(arr),
			any:  () => Promise.any(arr),
			race: () => Promise.race(arr),
		}
	},
});


class CustomElement extends HTMLElement {
	constructor() {
		super();

		this["--shadow"]  = this.attachShadow({mode: 'open'})
		this["--handler"] = {}
	}
	attributeChangedCallback(name, oldValue, newValue) {
		//  to use set follow to custom elements
		//
		//	static get observedAttributes() {
		//		return ["cluster"];
		//	}
		this.fireEvent("attribute-changed", {
			name: name,
			old: oldValue,
			new: newValue,
		});
	}
	connectedCallback()  {
		this.fireEvent("connected")
	}
	disconnectedCallback() {
		this.fireEvent("disconnected")
	}
	get shadow() {
		return this["--shadow"];
	}
	handler(h) {
		var name = h instanceof Function ? h.name : h;
		var f = h instanceof Function ? h : this[h];

		if (!this["--handler"][name]) {
			this["--handler"][name] = evt => f.call(this, evt.detail);
		}
		return this["--handler"][name];
	}
}

$.CustomElement = CustomElement;

export default $;
