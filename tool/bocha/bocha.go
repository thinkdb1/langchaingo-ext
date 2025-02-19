package bocha

import (
	"context"
	"errors"
	"github.com/thinkdb1/langchaingo-ext/tool/bocha/internal"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
	"os"
)

var ErrMissingToken = errors.New("missing the bocha API key, set it in the BOCHA_API_KEY environment variable")

type Tool struct {
	CallbacksHandler callbacks.Handler
	client           *internal.Client
}

var _ tools.Tool = Tool{}

// New creates a new bocha tool to search on internet.
func New(opts ...Option) (*Tool, error) {
	options := &options{
		apiKey: os.Getenv("BOCHA_API_KEY"),
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.apiKey == "" {
		return nil, ErrMissingToken
	}
	if options.count == 0 {
		options.count = 5
	}
	if options.count > 10 {
		options.count = 10
	}

	return &Tool{
		client: internal.New(options.apiKey, options.count, options.debug),
	}, nil
}

func (t Tool) Name() string {
	return "Bocha"
}

func (t Tool) Description() string {
	return `
	使用Bocha Web Search API 进行搜索互联网网页，输入应为搜索查询字符串，输出将返回搜索结果的详细信息，
包括网页标题、网页URL、网页摘要、网站名称、网站Icon、网页发布时间等。
`
}

func (t Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	result, err := t.client.Search(ctx, input)
	if err != nil {
		if errors.Is(err, internal.ErrNoGoodResult) {
			return "No Bocha Search Results was found", nil
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
