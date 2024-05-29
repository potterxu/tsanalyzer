package errinfo

import "errors"

var (
	ErrCellAlreadyStart    error = errors.New("cell already started")
	ErrCellAlreadyStop     error = errors.New("cell already stopped")
	ErrCellNotSupport      error = errors.New("cell not supported")
	ErrInvalidCellConfig   error = errors.New("invalid cell config")
	ErrInvalidMethod       error = errors.New("invalid method")
	ErrFailedToConnectCell error = errors.New("failed to connect cell")
	ErrFailedToBuildGraph  error = errors.New("failed to build graph")
)
