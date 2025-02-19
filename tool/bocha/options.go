package bocha

type options struct {
	apiKey string
	count  uint
	debug  bool
}

type Option func(*options)

// WithAPIKey passes the bocha API token to the client. If not set, the token
// is read from the BOCHA_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithCount set the number of search result
func WithCount(count uint) Option {
	return func(opts *options) {
		opts.count = count
	}
}

// WithDebug Enable debug mode
func WithDebug(debug bool) Option {
	return func(opts *options) {
		opts.debug = debug
	}
}
