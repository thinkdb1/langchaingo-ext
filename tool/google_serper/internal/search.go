package internal

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

const _searchUrl = "https://google.serper.dev/search"

type searchItem struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	Snippet  string `json:"snippet"`
	Date     string `json:"date"`
	Position int    `json:"position"`
}
type searchData struct {
	Content []searchItem `json:"organic"`
}

var _ searchInf = &searchData{}

func (r *searchData) OutputResponse() (string, error) {
	if len(r.Content) == 0 {
		return "", ErrNoGoodResult
	}
	webpages := r.Content
	formattedResults := ""
	for id, page := range webpages {
		formattedResults = formattedResults + fmt.Sprintf("(id: %d\ntitle: %s\nlink: %s\nsnippet: %s\ndate: %s\n)",
			id, page.Title, page.Link, page.Snippet, page.Date,
		)
	}
	return formattedResults, nil
}
func (r *searchData) HandleRequest(req *resty.Request) {
	req.SetResult(r)
}
func (r *searchData) GetUrl() string {
	return _searchUrl
}
