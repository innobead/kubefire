package validate

import (
	"fmt"
	intcmd "github.com/innobead/kubefire/internal/cmd"
	"github.com/innobead/kubefire/internal/di"
	interr "github.com/innobead/kubefire/internal/error"
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/pkg/errors"
	"regexp"
)

func CheckPrerequisites() error {
	if intcmd.CurrentPrerequisitesInfos().Matched() {
		return nil
	}

	return errors.WithMessage(interr.IncorrectRequiredPrerequisitesError, "check your installed prerequisites by `ignite info`, then install/update via 'kubefire install'")
}

func CheckClusterExist(name string) error {
	_, err := di.ConfigManager().GetCluster(name)
	if err != nil {
		return errors.WithMessage(interr.ClusterNotFoundError, Field("cluster", name))
	}

	return nil
}

func CheckNodeExist(name string) error {
	if _, err := di.NodeManager().GetNode(name); err != nil {
		return errors.WithMessage(interr.NodeNotFoundError, Field("node", name))
	}

	return nil
}

func CheckClusterVersion(version string) error {
	if version == "" {
		return nil
	}

	if matched, _ := regexp.MatchString(`^v\d+\.\d+(\.\d+)?$`, version); !matched {
		return errors.WithMessage(interr.ClusterVersionInvalidError, Field("version", version))
	}

	return nil
}

func CheckBootstrapperType(bootstrapper string) error {
	if !bootstrap.IsValid(bootstrapper) {
		return errors.WithMessage(interr.BootstrapperNotFoundError, Field("bootstrapper", bootstrapper))
	}

	return nil
}

func Field(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
