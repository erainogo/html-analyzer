package handlers

import (
	"errors"
	"net/http"
	"strings"
)

func getResponse(url string) (*http.Response, error) {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	// call the url
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New("unable to reach URL")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("received non-200 status: " + resp.Status)
	}

	return resp, nil
}
