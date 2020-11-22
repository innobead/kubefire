package cmd

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/innobead/kubefire/pkg/constants"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

type BootstrapperVersionInfo struct {
	Kubeadm string
	K3s     string
	RKE     string
}

func BootstrapperVersionInfos() *BootstrapperVersionInfo {
	wgDone := sync.WaitGroup{}
	wgDone.Add(3)

	versionsMap := map[string]string{}

	for _, bootstrapperType := range []string{constants.KUBEADM, constants.K3S, constants.RKE} {
		bootstrapperType := bootstrapperType
		go func() {
			versionsMap[bootstrapperType] = getVersions(bootstrapperType)
			wgDone.Done()
		}()
	}

	wgDone.Wait()

	return &BootstrapperVersionInfo{
		Kubeadm: versionsMap[constants.KUBEADM],
		K3s:     versionsMap[constants.K3S],
		RKE:     versionsMap[constants.RKE],
	}
}

func getVersions(bootstrapperType string) string {
	_, supportedVersions, err := bootstrap.GenerateSaveBootstrapperVersions(bootstrapperType, di.ConfigManager())
	if err != nil {
		logrus.WithError(err).Println()
		return ""
	}

	var versions []string
	for _, v := range supportedVersions {
		versions = append(versions, v.Display())
	}

	return strings.Join(versions, ",")
}
