package main

import (
	"errors"
	"syscall"
	"testing"
)

func TestMutexAlreadyExistsUsesDirectCreateMutexError(t *testing.T) {
	cases := []struct {
		err    error
		exists bool
	}{
		{syscall.Errno(183), true},
		{syscall.Errno(0), false},
		{syscall.Errno(5), false},
		{errors.New("arbitrary failure"), false},
		{nil, false},
		{errors.Join(errors.New("wrapped"), syscall.Errno(183)), true},
	}
	for _, tc := range cases {
		if got := mutexAlreadyExists(tc.err); got != tc.exists {
			t.Errorf("mutexAlreadyExists(%v)=%v, want %v", tc.err, got, tc.exists)
		}
	}
}
