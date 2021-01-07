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
	K0s     string
}

func BootstrapperVersionInfos() *BootstrapperVersionInfo {
	wgDone := sync.WaitGroup{}
	wgDone.Add(len(bootstrap.BuiltinTypes))

	versionsMap := map[string]string{}

	for _, bootstrapperType := range bootstrap.BuiltinTypes {
		bootstrapperType := bootstrapperType
		//go func() {
		versionsMap[bootstrapperType] = getVersions(bootstrapperType)
		//wgDone.Done()
		//}()
	}
	//wgDone.Wait()

	return &BootstrapperVersionInfo{
		Kubeadm: versionsMap[constants.KUBEADM],
		K3s:     versionsMap[constants.K3S],
		RKE:     versionsMap[constants.RKE],
		RKE2:    versionsMap[constants.RKE2],
		K0s:     versionsMap[constants.K0s],
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
