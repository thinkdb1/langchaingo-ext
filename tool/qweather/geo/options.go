package geo

type options struct {
	apiKey string //X-QW-Api-Key
	number uint
	lang   string //language
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
func WithNumber(count uint) Option {
	return func(opts *options) {
		opts.number = count
	}
}

func language(c string) Option {
	return func(opts *options) {
		opts.lang = c
	}
}

// WithDebug Enable debug mode
func WithDebug(debug bool) Option {
	return func(opts *options) {
		opts.debug = debug
	}
}
