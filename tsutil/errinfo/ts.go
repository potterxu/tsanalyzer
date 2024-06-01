package errinfo

import "errors"

var (
	// ErrNilFinish is returned when the finish function is nil
	ErrNilFinish = errors.New("finish function is nil")
)
