package logger

// HTTPRequest is specific to handling request information. It allows the routes to quickly append information such as
// status / latency / whether the response was cached or not to the log
type HTTPRequest interface {
	Logger

	// SetStatus sets the status in the log as the response status of the API
	SetStatus(status int)

	// SetCached sets the status of the response (whether the response that was sent to the client was cached or dynamic).
	SetCached(cached bool)

	// StartTimer starts the timer for latency checking
	StartTimer()

	// SetLatency sets the latency of the request. This requires StartTimer to be called first.
	SetLatency()
}
