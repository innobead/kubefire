package error

import "github.com/pkg/errors"

var (
	IncorrectRequiredPrerequisitesError = errors.New("incorrect required prerequisites")
	NotFoundError                       = errors.New("not found")
	NodeNotFoundError                   = errors.New("node not found")
	ClusterNotFoundError                = errors.New("cluster not found")
	ClusterVersionInvalidError          = errors.New("version is invalid, but v<major>.<minor> supported only")
	BootstrapperNotFoundError           = errors.New("bootstrapper not found")
)

func CheckErrors(errorFuncs ...func() error) error {
	for _, errFunc := range errorFuncs {
		if err := errFunc(); err != nil {
			return err
		}
	}

	return nil
}
