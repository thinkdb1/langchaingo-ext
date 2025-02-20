package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
)

const _url = "https://api.bochaai.com/v1/web-search"

var (
	ErrNoGoodResult = errors.New("no search results found")
)

type data struct {
	WebPages struct {
		Value []struct {
			Name            string `json:"name"`
			Url             string `json:"url"`
			Summary         string `json:"summary"`
			SiteName        string `json:"siteName"`
			SiteIcon        string `json:"siteIcon"`
			DateLastCrawled string `json:"dateLastCrawled"`
		} `json:"value"`
	} `json:"webPages"`
}
type BochaResp struct {
	Code int   `json:"code"`
	Data *data `json:"data"`
}
type Client struct {
	apiKey string
	count  uint
	debug  bool
}

func New(apiKey string, count uint, debug bool) *Client {
	return &Client{
		apiKey: apiKey,
		count:  count,
		debug:  debug,
	}
}

func (s *Client) Search(ctx context.Context, query string) (string, error) {
	r := resty.New().R().SetDebug(s.debug).SetContext(ctx)
	webRes := new(BochaResp)
	resp, err := r.SetHeader("Authorization", "Bearer "+s.apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]any{"query": query,
			"freshness": "noLimit",
			"summary":   true,
			"count":     s.count,
		}).SetResult(webRes).Post(_url)
	if err != nil {
		return "", fmt.Errorf("search in bocha api: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", errors.New("search in bocha api, status code: " + strconv.Itoa(resp.StatusCode()))
	}
	if webRes.Code != 200 {
		return "", errors.New("error from Bocha API: {" + resp.String() + "or 'Unknown error'}")
	}
	if webRes.Data == nil {
		return "", ErrNoGoodResult
	}
	webpages := webRes.Data.WebPages.Value
	formattedResults := ""
	for id, page := range webpages {
		if id == 0 {
			continue
		}
		formattedResults = formattedResults + fmt.Sprintf("(id: %d\ntitle: %s\nUrl: %s\nSummary: %s\nSiteName: %s\nDateLastCrawled: %s\n)",
			id, page.Name, page.Url, page.Summary, page.SiteName, page.DateLastCrawled,
		)
	}
	return formattedResults, nil
}
