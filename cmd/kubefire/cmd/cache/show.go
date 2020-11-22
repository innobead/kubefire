package cache

import (
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/cache"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows cache info",
	RunE: func(cmd *cobra.Command, args []string) error {
		var caches []*cache.Cache

		for _, c := range cache.DefaultManagers(di.NodeManager()) {
			cs, err := c.ListAll(false)
			if err != nil {
				return errors.WithMessage(err, "failed to get cache info")
			}

			caches = append(caches, cs...)
		}

		err := di.Output().Print(caches, []string{"Type", "Path", "Description"}, "")
		if err != nil {
			return errors.WithMessagef(err, "failed to print output of cache info")
		}

		return nil
	},
}

func init() {
	intcmd.AddOutputFlag(showCmd)
}
