package browse

import (
	"net/http"
	"net/url"
	"strings"
)

func addDefaultHeader(req *http.Request) {
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36 Edge/18.18363")
	req.Header.Add("Accept-Language", "ja")
	req.Header.Add("Connection", "Keep-Alive")
	req.Header.Add("Accept", "text/html, application/xhtml+xml, application/xml; q=0.9, */*; q=0.8")
}

func newGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	addDefaultHeader(req)
	return req, nil
}

func newPostFormRequest(url string, values url.Values) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	addDefaultHeader(req)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	return req, nil
}
