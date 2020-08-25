package kubeconfig

import (
	"fmt"
	"github.com/innobead/kubefire/internal/config"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/internal/validate"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var getCmd = &cobra.Command{
	Use:     "get [cluster-name]",
	Aliases: []string{"g"},
	Short:   "Get the kubeconfig of cluster",
	Args:    validate.OneArg("cluster name"),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.CheckClusterExist(args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.SetLevel(logrus.ErrorLevel)

		name := args[0]

		cluster, err := di.ClusterManager().Get(name)
		if err != nil {
			return errors.WithMessagef(err, "failed to get cluster (%s) info", name)
		}

		destDir := os.TempDir()
		defer func() {
			_ = os.RemoveAll(destDir)
		}()

		config.Bootstrapper = cluster.Spec.Bootstrapper
		di.DelayInit(true)

		wd, _ := os.Getwd()
		kubeconfigPath, err := di.Bootstrapper().DownloadKubeConfig(cluster, wd)
		if err != nil {
			return errors.WithMessagef(err, "failed to download kubeconfig of cluster (%s)", name)
		}

		bytes, err := ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return err
		}

		fmt.Print(string(bytes))

		return nil
	},
}
