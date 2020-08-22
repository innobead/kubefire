package util

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

func HttpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(body), nil
}
