package script

import (
	"fmt"
	"github.com/innobead/kubefire/pkg/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

type Type string

const (
	InstallPrerequisites   Type = "install-prerequisites.sh"
	UninstallPrerequisites Type = "uninstall-prerequisites.sh"

	InstallPrerequisitesKubeadm Type = "install-prerequisites-kubeadm.sh"
	InstallPrerequisitesSkuba   Type = "install-prerequisites-skuba.sh"
)

const (
	DownloadScriptEndpointFormat = "https://raw.githubusercontent.com/innobead/kubefire/master/scripts/%s"
)

func InstallPrerequisitesFile(version string) string {
	return path.Join(config.BinDir, version, string(InstallPrerequisites))
}

func UninstallPrerequisitesFile(version string) string {
	return path.Join(config.BinDir, version, string(UninstallPrerequisites))
}

func RemoteScriptUrl(script Type) string {
	return fmt.Sprintf(DownloadScriptEndpointFormat, script)
}

func Download(script Type, version string, force bool) error {
	log := logrus.WithFields(
		logrus.Fields{
			"version": version,
			"force":   force,
		},
	)
	log.Infof("downloading script (%s)", script)

	if version == "master" {
		log.Infof("changing to force download script (%s) because tag version is master", script)
		force = true
		log = log.WithField("force", force)
	}

	var err error

	switch script {
	case InstallPrerequisites:
		url := RemoteScriptUrl(InstallPrerequisites)
		destFile := InstallPrerequisitesFile(version)

		log.Infof("downloading %s to save %s", url, destFile)
		err = downloadScript(
			url,
			destFile,
			force,
		)

	case UninstallPrerequisites:
		url := RemoteScriptUrl(UninstallPrerequisites)
		destFile := UninstallPrerequisitesFile(version)

		log.Infof("downloading %s to save %s", url, destFile)
		err = downloadScript(
			url,
			destFile,
			force,
		)
	}

	if err != nil {
		return errors.WithMessagef(err, "failed to download script (%s)", script)
	}

	return nil
}

func Run(script Type, version string) error {
	log := logrus.WithFields(
		logrus.Fields{
			"version": version,
		},
	)
	log.Infof("running script (%s)", script)

	var err error

	switch script {
	case InstallPrerequisites:
		f := InstallPrerequisitesFile(version)

		log.Infof("running %s", f)
		err = runScript(f)

	case UninstallPrerequisites:
		f := UninstallPrerequisitesFile(version)

		log.Infof("running %s", f)
		err = runScript(f)
	}

	if err != nil {
		return errors.WithMessagef(err, "failed to run script (%s)", script)
	}

	return nil
}

func downloadScript(url string, destFile string, force bool) error {
	if _, err := os.Stat(destFile); os.IsExist(err) {
		if !force {
			return nil
		}

		if err := os.RemoveAll(destFile); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil && err != os.ErrExist {
		return errors.WithStack(err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return errors.WithStack(err)
	}
	defer resp.Body.Close()

	out, err := os.Create(destFile)
	if err != nil {
		return errors.WithStack(err)
	}
	defer out.Close()

	if err := out.Chmod(0755); err != nil {
		return errors.WithStack(err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func runScript(script string) error {
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return errors.WithStack(err)
	}

	cmd := exec.Command("sudo", script)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
