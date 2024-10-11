package types

import "errors"

var (
	ErrInvalidTargets   = errors.New("invalid or empty targets")
	ErrInvalidTemplates = errors.New("invalid or empty templates")
	ErrNoActiveHost     = errors.New("could not discovered active host")
	ErrNoExistPort      = errors.New("could not scanned exist port")
)

var (
	ErrHasBeenStopped          = errors.New("entry has been stopped")
	ErrAlreadyRunningOrStopped = errors.New("entry already running or stopped")
	ErrStoppedOrNotRunning     = errors.New("entry stopped or not running")
)
