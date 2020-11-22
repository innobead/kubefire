package cache

import (
	"github.com/innobead/kubefire/internal/di"
	"github.com/innobead/kubefire/pkg/cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"reflect"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Deletes caches",
	Aliases: []string{"rm", "del"},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, c := range cache.DefaultManagers(di.NodeManager()) {
			err := c.DeleteAll()
			if err != nil {
				logrus.Error(errors.WithMessagef(err, "failed to delete %s caches", reflect.TypeOf(c).Name()))
			}
		}

		return nil
	},
}
