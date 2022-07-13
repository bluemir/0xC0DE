import  "./v1/main.js";

let events = new EventSource("/stream");
events.on("message", function(evt) {
	console.log(evt);
})
events.onmessage = (evt) => {
	console.log(evt);
}
events.on("time", function(evt) {
	console.log(evt);
})
