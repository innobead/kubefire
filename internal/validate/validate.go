package validate

import (
	"fmt"
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	interr "github.com/innobead/kubefire/internal/error"
	"github.com/pkg/errors"
)

func RequiredPrerequisites() error {
	if intcmd.CurrentPrerequisitesInfos().Matched() {
		return nil
	}

	return errors.WithMessage(interr.IncorrectRequiredPrerequisitesError, "check your environment by `ignite info`")
}

func ClusterExist(name string) error {
	_, err := di.ConfigManager().GetCluster(name)
	if err != nil {
		return errors.WithMessage(interr.ClusterNotFoundError, field("cluster", name))
	}

	return nil
}

func field(key, value string) string {
	return fmt.Sprintf("%s = %s", key, value)
}
