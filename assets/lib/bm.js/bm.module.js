// bluemir's micro js library.
// light-weight & simple & vanilla friendly
//
// Usage
// import * as $ from "bm.module.js";
export let config = {
	hook: {
		preRequest: function(method, url, opt) { return opt }
	},
}

// global event listener
export let events = new EventTarget();

export function get(target, query) {
	if(target.querySelector instanceof Function) {
		return target.querySelector(query);
	}
	return document.querySelector(target)
}
export function all(target, query) {
	if(target.querySelectorAll instanceof Function) {
		return target.querySelectorAll(query);
	}
	return document.querySelectorAll(target);
}
export function create(tagname, attr = {}) {
	let newTag = document.createElement(tagname);
	if (attr.$text){
		newTag.appendChild(document.createTextNode(attr.$text));
	}
	if (attr.$html){
		newTag.innerHTML = attr.$html;
	}
	if (attr.$child) {
		if (attr.$child instanceof Array){
			attr.$child.forEach(n => newTag.appendChild(n))
		} else {
			newTag.appendChild(attr.$child)
		}
	}
	if (attr.$values) {
		Object.entries(attr.$values).forEach(([k, v]) => {
			newTag[k] = v;
		});
	}
	Object.entries(attr).filter(([key, values]) => key[0] != "$").forEach(([key, value]) => {
		newTag.setAttribute(key, value);
	});
	return newTag;
}
export async function request(method, url, options = {}) {
	try {
		var opts = config.hook.preRequest(method, url, options) || options;
	} catch(e) {
		var opts = options;
	}

	if (opts.timestamp === true) {
		opts.query = opts.query || {};
		opts.query["_timestamp"] = Date.now();
	}

	// parse url
	const u = new URL(url, location);
	opts.query = [...u.searchParams.entries()].reduce((obj, [key, value]) => {
		obj[key] = value
		return obj
	}, opts.query || {});

	u.search = "";
	url = u.href

	return new Promise(function(resolve, reject) {
		let req = new XMLHttpRequest();

		if (opts.timeout) {
			req.timeout = opts.timeout
		}

		req.addEventListener("readystatechange", function(){
			if (req.readyState == 4) {
				let result = {
					statusCode: req.status,
					text:       req.responseText,
					raw:        req.response,
				};

				let contentType = req.getResponseHeader("Content-Type") || "";
				if(contentType.includes("application/json")) {
					result.json = JSON.parse(result.text);
				}

				if (req.status >= 200 && req.status < 300){
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

		// set default accept
		req.setRequestHeader("Accept", "application/json,*/*");

		req.withCredentials = opts.withCredentials;
		Object.keys(opts.headers || {}).forEach(function(name){
			req.setRequestHeader(name, opts.headers[name]);
		});

		opts.body = opts.body || opts.data;

		switch (typeof opts.body) {
			case "object":
				if (opts.body instanceof FormData) {
					req.send(opts.body);
				} else {
					req.setRequestHeader("Content-Type", "application/json")
					req.send(JSON.stringify(opts.body))
				}
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
}
export async function timeout(ms) {
	return new Promise(function(resolve, reject){
		setTimeout(resolve, ms);
	});
}
export function defer() {
	let ret = {}
	ret.promise = new Promise(function(resolve, reject){
		ret.resolve = resolve;
		ret.reject = reject;
	});
	return ret;
}
export function prevent(func){
	return function(evt){
		evt.preventDefault();
		return func();
	}
}
export function form(form) {
	let fd = new FormData(form)
	return Array.from(fd).reduce((obj, [k, v] )=> {
		switch(get(form, `[name=${k}]`).attr("type")) {
			case "number":
				obj[k] = v-0;
				break;
			default:
				obj[k] = v;
				break;
		}
		return obj;
	}, {});
}
export function debounce(func, {timeout = 200} = {}) {
	let timer;

	return function(...args) {
		clearTimeout(timer);
		timer = setTimeout(_=> func.apply(this, args), timeout)
	}
}

// for await ( let dt of $.frames()){ /* do something */ }
export function frames({fps = 30} = {}) {
	let fpsInterval = 1000 / fps;
	let then = Date.now();

	async function* f() {
		while(true) {
			yield new Promise((resolve, reject) => {
				const animate = () => {
					let now = Date.now();
					let elapsed = now - then;

					if (elapsed > fpsInterval) {
						then = now - (elapsed%fpsInterval);

						resolve(elapsed - (elapsed%fpsInterval))
					} else {
						// wait next frame
						requestAnimationFrame(animate)
					}
				}
				requestAnimationFrame(animate)
			});
		}
	}
	return f();
}

export function animateFrame(callback, {fps = 30} = {}) {
	let stop = false;
	let fpsInterval = 1000 / fps;
	let then = Date.now();
	animate();

	function animate() {
		if (stop) {
			return;
		}
		requestAnimationFrame(animate);

		let now = Date.now();
		let elapsed = now - then;

		if (elapsed > fpsInterval) {
			then = now - (elapsed % fpsInterval);

			let ret = callback(elapsed - (elapsed%fpsInterval));
			if (ret && ret.stop) {
				stop = true;
			}
		}
	}
}
export function jq(data, query, value) {
	let keys = query.split("\\.").map(str => str.split(".")).reduce((p, c) => {
		if (p.length == 0 ) {
			return c;
		}
		let last = p.pop();
		let first = c.shift();

		return [].concat(p, [last+"."+first], c);
	});

	if (query[0] == ".") {
		keys.shift(); // remove first empty key
	}

	try {
		let visitor = data;
		while(keys.length > 1) {
			visitor = visitor[keys.shift()];
		}

		if (value !== undefined) {
			visitor[keys.shift()] = value;
			return value;
		} else {
			return visitor[keys.shift()];
		}
	} catch(e) {
		throw new ExtendedError("[$.jq] not found", e);
	}
}
export function merge(...args) {
	return args.reduce((target, src={}) => {
		return Object.entries(src).reduce((t, [key,value]) => {
			if (value instanceof Array) {
				return {
					...t,
					[key]: [...(t[key]||[]), ...value],
				}
			}
			if (value instanceof Object) {
				return {
					...t,
					[key]: merge(t[key], value)
				}
			}
			return {...t, [key]:value}
		}, target)
	}, {})
}

export function parsePathParam(pattern) {
	let ptn = pattern.split("/").filter( str => str.length > 0);

	let paths = location.pathname.split("/").filter( str => str.length > 0);

	return ptn.reduce((obj, current, index) => {
		if (obj === null) {
			return obj;
		}
		if (current.startsWith(":")) {
			let name = current.substring(1);
			obj[name] = paths[index]
		} else {
			if (current != paths[index]) {
				// not matched
				return null;
			}
		}
		return obj
	}, {});
}

class ExtendedError extends Error {
	constructor(message, error){
		super(message)

		this.name = error.name;

		this.cause = error;
		let message_lines = (this.message.match(/\n/g)||[]).length + 1;
		this.stack = this.stack.split('\n').slice(0, message_lines+1).join('\n') + '\n' + error.stack;
	}
}
export function wsURL (url){
	let u= new URL(url, document.location)
	u.protocol = document.location.protocol.includes("https") ? "wss:" : "ws:"
	return u;
}

export const util = {
	filter: {
		notNull: e => e != null,
		unique: (value, index, self) => self.indexOf(value) === index,
	},
	reduce: {
		appendChild: function(parent, child) {
			parent.appendChild(child);
			return parent;
		},
	},
};

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
		return key + "=" + encodeURIComponent(obj[key]);
	}).join("&");
}

Object.keyValues= function(obj, f) {
	return Object.entries(obj).map(([key, value]) => {
		return {key, value};
	});
}
Object.map = function(obj, f) {
	return Object.entries(obj).map(([key, value]) => f({key,value})).reduce((obj, {key,value}={}) => (key?{ ...obj, [key]: value}:obj), {});
}
Object.same = function(x, y) {
	if (x === null || x === undefined || y === null || y === undefined) {
		return x === y;
	}
	if (x.constructor !== y.constructor) {
		return false;
	}
	if (x instanceof RegExp || x instanceof Function) {
		return x === y;
	}
	if (x === y || x.valueOf() === y.valueOf()) {
		return true;
	}
	if (Array.isArray(x) && x.length !== y.length) {
		return false;
	}

	// if they are dates, they must had equal valueOf
	if (x instanceof Date) {
		return false;
	}

	if (!(x instanceof Object)) {
		return false;
	}
	if (!(y instanceof Object)) {
		return false;
	}
	let xk = Object.keys(x);
	let yk = Object.keys(y);

	if (xk.length != yk.length) {
		return false
	}
	if (!xk.every(i => yk.indexOf(i) !== -1)) {
		return false
	}

	// recursive object equality check
	return xk.every(i => Object.same(x[i], y[i]))
}

const sig = "__bm.js_inserted__"
function extend(TargetClass, proto){
	if (TargetClass.hasOwnProperty(sig)) {
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

	TargetClass[sig] = true
}
extend(Node, {
	appendTo: function(target) {
		target.appendChild(this);
		return this;
	},
	clear : function(filter) {
		let f = filter || function(e) { return true };
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
	on: function(name, handler, opts) {
		this.addEventListener(name, handler, opts);
		return this;
	},
	off: function(name, handler, opts) {
		this.removeEventListener(name, handler, opts)
		return this;
	},
	fireEvent: function(name, detail) {
		let evt = new CustomEvent(name, {detail: detail});
		this.dispatchEvent(evt);
		return this;
	}
});

extend(NodeList, {
	map:    Array.prototype.map,
	filter: Array.prototype.filter,
	//"forEach": Array.prototype.forEach,
});
extend(HTMLCollection, {
	map:     Array.prototype.map,
	filter:  Array.prototype.filter,
	forEach: Array.prototype.forEach,
});

extend(Array, {
	unique: function(isSame) {
		if (!isSame) {
			return [... new Set(this)];
		}
		return this.filter((v, i)  => this.first(v, isSame) == i);
	},
	promise: function() {
		let arr = this;
		return {
			all:  () => Promise.all(arr),
			any:  () => Promise.any(arr),
			race: () => Promise.race(arr),
		}
	},
	first: function(v, isSame = ((a,b)=>a==b)) {
		for (let i = 0; i < this.length; i ++){
			if(isSame(this[i], v)) {
				return i;
			}
		}
	},
});

extend(HTMLElement, {
	// syntactic sugars
	connectedCallback() {
		this.render && this.render();
		this.onConnected && this.onConnected();
		this.fireEvent("connected");
	},
	disconnectedCallback() {
		this.onDisconnected && this.onDisconnected();
		this.fireEvent("disconnected")
	},
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
		this.onAttributeChanged(name, oldValue, newValue);
	},
	onAttributeChanged() {
		this.render && this.render();
	},
})


export class CustomElement extends HTMLElement {
	constructor({enableShadow = true} = {}) {
		super();

		if (enableShadow) {
			// this.shadowRoot
			this.attachShadow({mode: 'open'})
		}
	}
}


export class AwaitEventTarget {
	constructor() {
		this.handlers = new Map();
	}

	// method
	addEventListener(eventName, handler) {
		if (!this.handlers.has(eventName)) {
			this.handlers.set(eventName, new Set());
		}
		this.handlers.get(eventName).add(handler);
	}
	removeEventListener(eventName, handler) {
		if (!this.handlers.has(eventName)) {
			return;
		}
		this.handlers.get(eventName).delete(handler);
	}
	dispatchEvent(event) {
		let name = event.type;
		if (!this.handlers.has(name)) {
			return;
		}
		return [...this.handlers.get(name)].map(handler => {
			return handler(event);
		}).promise().all();
	}

	// syntactic sugar
	on(eventName, handler, opts) {
		this.addEventListener(eventName, handler, opts);
		return this;
	}
	off(eventName, handler, opts) {
		this.removeEventListener(eventName, handler, opts);
		return this;
	}
	fireEvent(name, detail) {
		let evt = new CustomEvent(name, {detail: detail});
		// name will be evt.type
		return this.dispatchEvent(evt);
	}
}

export class AwaitQueue {
	constructor() {
		this.queue = [];
		this.resolve = null;
	}
	[Symbol.iterator]() {
		let next = () => {
			if (this.queue.length > 0) {
				return {
					value: this.queue.shift(),
				}
			}
			return {
				value: () => new Promise((resolve) => {
					this.resolve = resolve;
				}),
			};
		}
		return { next }
	}
	add(f) {
		if (!(f instanceof Function)) {
			throw Error("must put function");
		}
		if(this.resolve) {
			this.resolve(f());
			this.resolve = null;
			return
		}
		this.queue.push(f)
	}
	remove(f) {
		this.queue = this.queue.filter(job => job != f);
	}
	get length() {
		return this.queue.length;
	}
}

