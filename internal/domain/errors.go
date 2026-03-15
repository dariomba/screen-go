package domain

import "errors"

var ErrJobNotFound = errors.New("job not found")

type JobInvalidError struct {
	Message string
}

func (e *JobInvalidError) Error() string {
	return e.Message
}
