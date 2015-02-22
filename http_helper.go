package main

import (
	"net/http"
	"net/url"
	"strings"
)

func post(url string, header http.Header, values url.Values) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(values.Encode()))
	if err != nil {
		return
	}

	req.Header = header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err = client.Do(req)
	return

}
