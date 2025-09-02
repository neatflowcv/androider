package virtualmachine

import "errors"

var (
	ErrInstanceNotFound       = errors.New("instance not found")
	ErrInstanceAlreadyRunning = errors.New("instance already running")
)
