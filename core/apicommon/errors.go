package apicommon

import (
	"errors"
	"fmt"
)

type UserFacingError struct {
	message string
}

func (e *UserFacingError) Error() string {
	return e.message
}

func NewUserFacingError(format string, args ...any) error {
	full := fmt.Sprintf(format, args...)
	return &UserFacingError{
		message: full,
	}
}

func SanitizeError(err error, placeholder string) string {
	if e, ok := errors.AsType[*UserFacingError](err); ok {
		return e.Error()
	}

	return placeholder
}
