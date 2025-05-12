package properties

import "fmt"

type PropertyNotFoundError struct {
	Key string
}

func (e *PropertyNotFoundError) Error() string {
	return fmt.Sprintf("property not found: %s", e.Key)
}

func NewPropertyNotFoundError(key string) error {
	return &PropertyNotFoundError{Key: key}
}

func IsPropertyNotFoundError(err error) bool {
	_, ok := err.(*PropertyNotFoundError)
	return ok
}
