package main

import (
	"errors"
	"syscall"
)

// ERROR_ALREADY_EXISTS (183) is the only indication that another instance
// already created our named mutex. The error must come from CreateMutexW itself.
func mutexAlreadyExists(createErr error) bool {
	return errors.Is(createErr, syscall.Errno(183))
}
