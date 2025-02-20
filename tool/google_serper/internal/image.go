package internal

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

const _imageUrl = "https://google.serper.dev/images"

type ImageItem struct {
	Title           string `json:"title"`
	Link            string `json:"link"`
	Snippet         string `json:"snippet"`
	Date            string `json:"date"`
	Position        int    `json:"position"`
	ImageUrl        string `json:"imageUrl"`
	ImageWidth      int    `json:"imageWidth"`
	ImageHeight     int    `json:"imageHeight"`
	ThumbnailUrl    string `json:"thumbnailUrl"`
	ThumbnailWidth  int    `json:"thumbnailWidth"`
	ThumbnailHeight int    `json:"thumbnailHeight"`
}

type imageData struct {
	Content []ImageItem `json:"images"`
}

var _ searchInf = &imageData{}

func (r *imageData) OutputResponse() (string, error) {
	if len(r.Content) == 0 {
		return "", ErrNoGoodResult
	}
	webpages := r.Content
	formattedResults := ""
	for id, page := range webpages {
		formattedResults = formattedResults + fmt.Sprintf("(id: %d\ntitle: %s\nlink: %s\nsnippet: %s\ndate: %s\n"+
			"position: %d\nimageUrl: %s\nimageWidth: %d\nimageHeight: %d\nthumbnailUrl: %s\nthumbnailWidth: %d\nthumbnailHeight: %d\n)",
			id, page.Title, page.Link, page.Snippet, page.Date,
			page.Position, page.ImageUrl, page.ImageWidth, page.ImageHeight,
			page.ThumbnailUrl, page.ThumbnailWidth, page.ThumbnailHeight,
		)
	}
	return formattedResults, nil
}
func (r *imageData) HandleRequest(req *resty.Request) {
	req.SetResult(r)
}
func (r *imageData) GetUrl() string {
	return _imageUrl
}
