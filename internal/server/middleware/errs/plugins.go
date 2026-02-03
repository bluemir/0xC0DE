package errs

// formatAdjusters is a list of functions that can override the response format
// based on error type. Each adjuster receives the current format and error,
// and returns the (possibly modified) format.
var formatAdjusters = []func(string, error) string{
	//adjustForOAuth,
}
