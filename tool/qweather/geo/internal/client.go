package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/url"
	"strconv"
	"strings"
)

const _url = "https://geoapi.qweather.com/v2/city/lookup?"

var (
	ErrNoGoodResult = errors.New("no search results found")
)

type Resp struct {
	Code     string     `json:"code"`
	Location []Location `json:"location"`
	Error    *Errors    `json:"error"`
}
type Errors struct {
	Status int    `json:"status"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}
type Location struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Adm2    string `json:"adm2"`
	Adm1    string `json:"adm1"`
	Country string `json:"country"`
	Tz      string `json:"tz"`
}

type Client struct {
	apiKey string
	debug  bool
	number uint
	lang   string //language
}

func New(apiKey string, debug bool, number uint, lang string) *Client {
	return &Client{
		apiKey: apiKey,
		debug:  debug,
		number: number,
		lang:   lang, //language
	}
}

func (s *Client) Search(ctx context.Context, query string) (string, error) {
	qList := strings.Split(query, "/")
	if len(qList) != 2 {
		return "", errors.New("query Must follow province/city  like 河北省/石家庄市")
	}
	requestUri := fmt.Sprintf(_url+"location=%s&adm=%s&range=%s&number=%d", url.QueryEscape(qList[1]), url.QueryEscape(qList[0]), "cn", s.number)
	if s.lang != "" {
		requestUri += "&lang=" + s.lang
	}
	r := resty.New().R().SetDebug(s.debug).SetContext(ctx)
	webRes := new(Resp)
	resp, err := r.SetHeader("X-QW-Api-Key", s.apiKey).SetHeader("User-Agent", "ext").
		Get(requestUri)
	if err != nil {
		return "", fmt.Errorf("search in q-geo api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", errors.New("search in q-geo api, status code: " + strconv.Itoa(resp.StatusCode()))
	}
	if webRes.Error != nil {
		return "", errors.New("error from q-geo API: {" + webRes.Error.Title + ":" + webRes.Error.Detail + "}")
	}
	if webRes.Code != "200" {
		return "", errors.New("error from q-geo API: {" + resp.String() + "or 'Unknown error'}")
	}
	if webRes.Location == nil {
		return "", ErrNoGoodResult
	}
	buf := new(strings.Builder)
	if err := json.NewEncoder(buf).Encode(webRes.Location); err != nil {
		return "", err
	}
	return buf.String(), nil
}
