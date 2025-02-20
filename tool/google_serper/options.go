package google_serper

type options struct {
	apiKey     string
	debug      bool
	searchType uint
}

type Option func(*options)

// WithAPIKey passes the google_serper API token to the client. If not set, the token
// is read from the GOOGLE_SERPER_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithDebug Enable debug mode
func WithDebug(debug bool) Option {
	return func(opts *options) {
		opts.debug = debug
	}
}

// WithSearchType select the type for search ,default 1
func WithSearchType(k uint) Option {
	return func(opts *options) {
		opts.searchType = k
	}
}
