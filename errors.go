package main

import (
	"errors"
	"fmt"
)

var (
	ErrValidation    = errors.New("Validation failed")
	ErrUnsupportedOs = errors.New("This OS is not supported")
	ErrSignal        = errors.New("Recieved a signal")
)

type stepErr struct {
	step  string
	msg   string
	cause error
}

func (s *stepErr) Error() string {
	return fmt.Sprintf("Step: %q: %s: Cause: %v", s.step, s.msg, s.cause)
}

func (s *stepErr) Is(tatget error) bool {
	t, ok := tatget.(*stepErr)
	if !ok {
		return false
	}

	return t.step == s.step
}

func (s *stepErr) Unwrap() error {
	return s.cause
}
