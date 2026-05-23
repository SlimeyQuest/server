package network

import "errors"

var (
	errUnknownMessage = errors.New("unknown gameplay message")
	errInvalidSession = errors.New("invalid session")
)
