package firecracker

import (
	"encoding/json"
	"github.com/buger/jsonparser"
	resty "github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type Interface interface {
	Precheck() error
	Download() (filepath string, err error)
	Install() error
}

type Ignite struct {
	Version  string
	filepath string
}

func (i *Ignite) Precheck() error {
	logrus.Info("Pre-checking the perquisites of ignite")

	// https://github.com/weaveworks/ignite/blob/master/docs/installation.md
	return nil
}

func (i *Ignite) Download() (filepath string, err error) {
	logrus.Infof("Download ignite (%s", i.Version)

	//export VERSION=v0.6.3
	//export GOARCH=$(go env GOARCH 2>/dev/null || echo "amd64")
	//
	//for binary in ignite ignited; do
	//echo "Installing ${binary}..."
	//curl -sfLo ${binary} https://github.com/weaveworks/ignite/releases/download/${VERSION}/${binary}-${GOARCH}
	//chmod +x ${binary}
	//sudo mv ${binary} /usr/local/bin
	//done

	request := resty.New().R()

	response, err := request.Get("")
	if err != nil {
		return "", err
	}

	jsonparser.Get(response.Body(), "")

	data := make(map[string]interface{})
	if err = json.Unmarshal(response.Body(), &data); err != nil {
		return "", nil
	}

	return filepath, nil
}

func (i *Ignite) Install(filePath string) error {
	logrus.Infof("Download ignite (%s)", i.Version)

	return nil
}
