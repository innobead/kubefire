package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type ImageType = string

type ImageInfo struct {
	Image string
	Type  string
}

var (
	rootFsImage = "image"
	kernelImage = "kernel"
)

func ImageInfos() (*[]ImageInfo, error) {
	var images []ImageInfo

	for _, t := range []string{rootFsImage, kernelImage} {
		infos, err := imageInfos(t)
		if err != nil {
			return nil, err
		}

		images = append(images, *infos...)
	}

	return &images, nil
}

func imageInfos(imageType ImageType) (*[]ImageInfo, error) {
	result, err := http.DefaultClient.Get(fmt.Sprintf("https://raw.githubusercontent.com/innobead/kubefire/master/generated/%s.list", imageType))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	bytes, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var images []ImageInfo
	for _, img := range strings.Split(strings.TrimRight(string(bytes), "\n"), "\n") {
		images = append(images, ImageInfo{
			Image: img,
			Type:  imageTypeString(imageType),
		})
	}

	return &images, nil
}

func imageTypeString(imgType ImageType) string {
	switch imgType {
	case rootFsImage:
		return "RootFS"
	default:
		return "Kernel"
	}
}
