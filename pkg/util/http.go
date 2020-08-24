package util

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

func HttpGet(url string) (string, *http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", resp, errors.WithStack(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", resp, errors.WithStack(err)
	}

	return string(body), resp, nil
}
