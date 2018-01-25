package remote

import (
	"errors"
)

var (
	// ErrNotExist record is not exist error
	ErrNotExist = errors.New("Not exist error")
	// ErrRequestCall record is not request error
	ErrRequestCall = errors.New("Request call error")
	// ErrTimeOut record timeout error
	ErrTimeOut = errors.New("Remote call error")
)
