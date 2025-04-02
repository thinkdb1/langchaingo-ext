package weather

import (
	"context"
	"errors"
	"github.com/thinkdb1/langchaingo-ext/tool/qweather/weather/internal"
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
	if Opts.unit == "" {
		Opts.unit = "m"
	}

	return &Tool{
		client: internal.New(Opts.apiKey, Opts.debug, Opts.unit, Opts.lang),
	}, nil
}

func (t Tool) Name() string {
	return "q-weather"
}

func (t Tool) Description() string {
	return `
	使用q-weather API 进行城市天气搜索搜索，输入：longitude,latitude，经纬度间使用“,”分隔。
	返回：天气信息列表,
	[{
		fxDate: 日期,
		sunrise: 日升时间,
		sunset: 日落时间,
		tempMax: 最高温度,
		tempMin: 最低温度,
		textDay: 白天天气状况文字描述，包括阴晴雨雪等天气状态的描述,
		textNight: 晚间天气状况文字描述，包括阴晴雨雪等天气状态的描述,
		windDirDay: 白天风向
		windScaleDay: 白天风力等级
		windDirNight:夜间当天风向
		windScaleNight: 夜间风力等级
		humidity: 相对湿度，百分比数值
		vis: 能见度，默认单位：公里
		uvIndex: 紫外线强度指数   		
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
			return "No q-weather Search Results was found", nil
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
