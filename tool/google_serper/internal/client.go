package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
)

const (
	TYPE_SEARCH = 1
	TYPE_IMAGES = 2
	TYPE_NEWS   = 3
)

var TypeMap = map[uint]string{
	TYPE_SEARCH: "Search",
	TYPE_IMAGES: "Images",
	TYPE_NEWS:   "News",
}
var (
	ErrNoGoodResult = errors.New("no search results found")
)

type searchInf interface {
	HandleRequest(req *resty.Request)
	OutputResponse() (string, error)
	GetUrl() string
}
type Client struct {
	apiKey string
	debug  bool
	inf    searchInf
}

func New(apiKey string, searchType uint, debug bool) *Client {
	c := &Client{
		apiKey: apiKey,
		debug:  debug,
	}
	switch searchType {
	case TYPE_SEARCH:
		c.inf = new(searchData)
	case TYPE_IMAGES:
		c.inf = new(imageData)
	case TYPE_NEWS:
		c.inf = new(NewsData)
	default:
		c.inf = new(searchData)
	}
	return c
}

func (s *Client) Search(ctx context.Context, query string) (string, error) {
	r := resty.New().R().SetDebug(s.debug).SetContext(ctx)
	s.inf.HandleRequest(r)
	resp, err := r.SetHeader("X-API-KEY", s.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{
			"q": query,
		}).Post(s.inf.GetUrl())
	if err != nil {
		return "", fmt.Errorf("search in google_serper api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", errors.New("search in google_serper api, status code: " + strconv.Itoa(resp.StatusCode()))
	}
	return s.inf.OutputResponse()
}
