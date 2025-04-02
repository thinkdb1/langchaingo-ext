package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
	"strings"
)

const _url = "https://devapi.qweather.com/v7/weather/7d?"

var (
	ErrNoGoodResult = errors.New("no search results found")
)

type Resp struct {
	Code  string      `json:"code"`
	Daily []DailyItem `json:"daily"`
	Error *Errors     `json:"error"`
}
type Errors struct {
	Status int    `json:"status"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type DailyItem struct {
	FxDate  string `json:"fxDate"`
	Sunrise string `json:"sunrise"`
	Sunset  string `json:"sunset"`
	//Moonrise      string `json:"moonrise"`
	//Moonset       string `json:"moonset"`
	//MoonPhase     string `json:"moonPhase"`
	//MoonPhaseIcon string `json:"moonPhaseIcon"`
	TempMax string `json:"tempMax"`
	TempMin string `json:"tempMin"`
	//IconDay       string `json:"iconDay"`
	TextDay string `json:"textDay"`
	//IconNight     string `json:"iconNight"`
	TextNight string `json:"textNight"`
	//Wind360Day    string `json:"wind360Day"`
	WindDirDay   string `json:"windDirDay"`
	WindScaleDay string `json:"windScaleDay"`
	//WindSpeedDay   string `json:"windSpeedDay"`
	//Wind360Night   string `json:"wind360Night"`
	WindDirNight   string `json:"windDirNight"`
	WindScaleNight string `json:"windScaleNight"`
	//WindSpeedNight string `json:"windSpeedNight"`
	Humidity string `json:"humidity"`
	Precip   string `json:"precip"`
	Pressure string `json:"pressure"`
	Vis      string `json:"vis"`
}

type Client struct {
	apiKey string
	debug  bool
	uint   string
	lang   string //language
}

func New(apiKey string, debug bool, uint string, lang string) *Client {
	return &Client{
		apiKey: apiKey,
		debug:  debug,
		uint:   uint,
		lang:   lang, //language
	}
}

func (s *Client) Search(ctx context.Context, query string) (string, error) {
	qList := strings.Split(query, ",")
	if len(qList) != 2 {
		return "", errors.New("query Must follow 'longitude/latitude' ；Longitude and latitude are floating point numbers ；like 116.41,39.92")
	}
	lon, err := strconv.ParseFloat(qList[0], 64)
	if err != nil {
		return "", errors.New("query Must follow 'longitude/latitude' ；Longitude and latitude are floating point numbers ； like 116.41,39.92")
	}
	lat, err := strconv.ParseFloat(qList[1], 64)
	if err != nil {
		return "", errors.New("query Must follow 'longitude/latitude' ；Longitude and latitude are floating point numbers ； like 116.41,39.92")
	}
	requestUri := fmt.Sprintf(_url+"location=%.2f,%.2f&unit=%s", lon, lat, s.uint)
	if s.lang != "" {
		requestUri += "&lang=" + s.lang
	}
	r := resty.New().R().SetDebug(s.debug).SetContext(ctx)
	webRes := new(Resp)
	resp, err := r.SetHeader("X-QW-Api-Key", s.apiKey).
		SetHeader("User-Agent", "ext").
		SetResult(webRes).
		Get(requestUri)
	if err != nil {
		return "", fmt.Errorf("search in q-weather api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", errors.New("search in q-weather api, status code: " + strconv.Itoa(resp.StatusCode()))
	}

	if webRes.Error != nil {
		return "", errors.New("error from q-weather API: {" + webRes.Error.Title + ":" + webRes.Error.Detail + "}")
	}
	if webRes.Code != "200" {
		return "", errors.New("error from q-weather API: {" + resp.String() + "or 'Unknown error'}")
	}
	if webRes.Daily == nil {
		return "", ErrNoGoodResult
	}
	buf := new(strings.Builder)
	if err := json.NewEncoder(buf).Encode(webRes.Daily); err != nil {
		return "", err
	}
	return buf.String(), nil
}
