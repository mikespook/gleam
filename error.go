package gleam

import (
	"errors"
)

var (
	ErrLoadScript = errors.New("Loading script error")
	ErrConnection = errors.New("Connection error")
	ErrParam      = errors.New("Params error")
)

type ErrorHandlerFunc func(error)
