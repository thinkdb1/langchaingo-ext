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
	Use the q-weather API to perform a city weather search.
Input is a json Serialized multiple strings: .
"{\"longitude\":10.111,\"latitude\":10.111,\"city\":\"city name for search\"}"

Return: A list of weather information
	[
  {
    "fxDate": "Date",
    "sunrise": "Sunrise time",
    "sunset": "Sunset time",
    "tempMax": "Maximum temperature",
    "tempMin": "Minimum temperature",
    "textDay": "Daytime weather description, including conditions such as sunny, cloudy, rainy, snowy, etc.",
    "textNight": "Nighttime weather description, including conditions such as sunny, cloudy, rainy, snowy, etc.",
    "windDirDay": "Daytime wind direction",
    "windScaleDay": "Daytime wind scale level",
    "windDirNight": "Nighttime wind direction",
    "windScaleNight": "Nighttime wind scale level",
    "humidity": "Relative humidity in percentage",
    "vis": "Visibility in kilometers (default unit)",
    "uvIndex": "UV index"
  }
]
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
