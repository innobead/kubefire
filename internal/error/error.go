package error

import "github.com/pkg/errors"

var (
	IncorrectRequiredPrerequisitesError = errors.New("incorrect required prerequisites")
	ClusterNotFoundError                = errors.New("cluster not found")
	NodeNotFoundError                   = errors.New("node not found")
)

func CheckErrors(errorFuncs ...func() error) error {
	for _, errFunc := range errorFuncs {
		if err := errFunc(); err != nil {
			return err
		}
	}

	return nil
}
