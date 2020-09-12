package cmd

import (
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ImageCmd = &cobra.Command{
	Use:     "image",
	Aliases: []string{"i"},
	Short:   "Show supported RootFS and Kernel images",
	PreRun: func(cmd *cobra.Command, args []string) {
		logrus.SetLevel(logrus.ErrorLevel)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		infos, err := intcmd.ImageInfos()
		if err != nil {
			return errors.WithMessage(err, "failed to print output of images info")
		}

		if err := di.Output().Print(infos, nil, ""); err != nil {
			return errors.WithMessage(err, "failed to print output of images info")
		}

		return nil
	},
}
