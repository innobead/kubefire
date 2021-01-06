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
	RKE2    string
}

func BootstrapperVersionInfos() *BootstrapperVersionInfo {
	bootstrapTypes := []string{constants.KUBEADM, constants.K3S, constants.RKE, constants.RKE2}

	wgDone := sync.WaitGroup{}
	wgDone.Add(len(bootstrapTypes))

	versionsMap := map[string]string{}

	for _, bootstrapperType := range bootstrapTypes {
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
		RKE2:    versionsMap[constants.RKE2],
	}
}

func getVersions(bootstrapperType string) string {
	_, supportedVersions, err := bootstrap.GenerateSaveBootstrapperVersions(bootstrapperType, di.ConfigManager())
	if err != nil {
		logrus.WithError(err).Errorln()
		return ""
	}

	var versions []string
	for _, v := range supportedVersions {
		versions = append(versions, v.Display())
	}

	return strings.Join(versions, ",")
}
