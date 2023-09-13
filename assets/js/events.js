let events = new EventSource("/stream");
events.on("message", evt => {
	console.log(evt);
})
events.on("time", evt => {
	console.log(evt);
})
events.on("error", evt => {
	events.close();
})
