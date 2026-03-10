package usecase

import "github.com/rafaelblt/go-auth/internal/shared"

type ErrorsMap map[error]error

func MapErrors(errs []error, errMap ErrorsMap) ([]error, []error) {
	mappeds := shared.NewSet[error]()
	unexpecteds := []error{}
	for _, err := range errs {
		expected, exists := MapError(err, errMap)
		if !exists {
			if err != nil {
				unexpecteds = append(unexpecteds, err)
			}
			continue
		}
		if mappeds.Contains(expected) {
			continue
		}
		mappeds.Add(expected)
	}
	return mappeds.Values(), unexpecteds
}

func MapError(err error, errMap ErrorsMap) (error, bool) {
	expected, exists := errMap[err]
	return expected, exists
}
