package internal

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

const _newsUrl = "https://google.serper.dev/news"

type NewsItem struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	Snippet  string `json:"snippet"`
	Date     string `json:"date"`
	Source   string `json:"source"`
	ImageUrl string `json:"imageUrl"`
	Position int    `json:"position"`
}
type NewsData struct {
	Content []NewsItem `json:"news"`
}

var _ searchInf = &NewsData{}

func (r *NewsData) OutputResponse() (string, error) {
	if len(r.Content) == 0 {
		return "", ErrNoGoodResult
	}
	webpages := r.Content
	formattedResults := ""
	for id, page := range webpages {
		formattedResults = formattedResults + fmt.Sprintf("(id: %d\ntitle: %s\nlink: %s\nsnippet: %s\ndate: %s\n"+
			"source: %s\nimageUrl: %s\n)",
			id, page.Title, page.Link, page.Snippet, page.Date,
			page.Source, page.ImageUrl,
		)
	}
	return formattedResults, nil
}
func (r *NewsData) HandleRequest(req *resty.Request) {
	req.SetResult(r)
}
func (r *NewsData) GetUrl() string {
	return _newsUrl
}
