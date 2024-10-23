function runEventhandler() {
	let events = new EventSource("/api/v1/stream");

	events.on("message", evt => {
		console.log(evt);
	})
	events.on("time", evt => {
		console.log(evt);
	})
	events.on("error", evt => {
		events.close();
	})
}
