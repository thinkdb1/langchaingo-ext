package geo

import (
	"context"
	"errors"
	"github.com/thinkdb1/langchaingo-ext/tool/qweather/geo/internal"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
	"os"
)

var ErrMissingToken = errors.New("missing the qweather API key, set it in the QWEATHER_API_KEY environment variable")

type Tool struct {
	CallbacksHandler callbacks.Handler
	client           *internal.Client
}

var _ tools.Tool = Tool{}

// New creates a new bocha tool to search on internet.
func New(opts ...Option) (*Tool, error) {
	Opts := &options{
		apiKey: os.Getenv("QWEATHER_API_KEY"),
	}

	for _, opt := range opts {
		opt(Opts)
	}

	if Opts.apiKey == "" {
		return nil, ErrMissingToken
	}
	if Opts.number == 0 {
		Opts.number = 5
	}
	if Opts.number > 10 {
		Opts.number = 10
	}

	return &Tool{
		client: internal.New(Opts.apiKey, Opts.debug, Opts.number, Opts.lang),
	}, nil
}

func (t Tool) Name() string {
	return "q-geo"
}

func (t Tool) Description() string {
	return `
	使用q-geo API 进行城市经纬度搜索搜索，输入：国家/省份/城市，返回：经纬度信息列表,
	[{
      "name":"城市",
      "lat":"纬度",
      "lon":"经度",
      "adm2":"所属上级行政区划",
      "adm1":"所属一级行政区域",
      "country":"国家",
      "tz":"所在时区",
}]
`
}

func (t Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	result, err := t.client.Search(ctx, input)
	if err != nil {
		if errors.Is(err, internal.ErrNoGoodResult) {
			return "No q-geo Search Results was found", nil
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
