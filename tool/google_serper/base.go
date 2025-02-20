package google_serper

import (
	"context"
	"errors"
	"github.com/thinkdb1/langchaingo-ext/tool/google_serper/internal"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
	"os"
)

var ErrMissingToken = errors.New("missing the google_serper API key, set it in the GOOGLE_SERPER_API_KEY environment variable")

type Tool struct {
	CallbacksHandler  callbacks.Handler
	client            *internal.Client
	DescriptionCustom string
}

var _ tools.Tool = Tool{}

// New creates a new google_serper tool to search on internet.
func New(opts ...Option) (*Tool, error) {
	options := &options{
		apiKey: os.Getenv("GOOGLE_SERPER_API_KEY"),
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.searchType == 0 {
		options.searchType = 1
	}

	if _, ok := internal.TypeMap[options.searchType]; !ok {
		return nil, errors.New("unknown search type")
	}

	if options.apiKey == "" {
		return nil, ErrMissingToken
	}

	return &Tool{
		client: internal.New(options.apiKey, options.searchType, options.debug),
	}, nil
}

func (t Tool) Name() string {
	return "google_serper"
}

func (t Tool) Description() string {
	if t.DescriptionCustom != "" {
		return t.DescriptionCustom
	}
	return `
	"A wrapper around Google Search. "
	"Useful for when you need to answer questions about current events. "
	"Always one of the first options when you need to find information on internet"
	"Input should be a search query."
`
}

func (t Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	result, err := t.client.Search(ctx, input)
	if err != nil {
		if errors.Is(err, internal.ErrNoGoodResult) {
			return "No google serper Search Results was found", nil
		}

		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}

		return "", err
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}
