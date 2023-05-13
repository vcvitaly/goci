package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type timeoutStep struct {
	step
	timeout time.Duration
}

func newTimeoutStep(name, exe, message, proj string, args []string, timeout time.Duration) timeoutStep {
	s := timeoutStep{}

	s.step = newStep(name, exe, message, proj, args)
	s.timeout = timeout
	if s.timeout == 0 {
		s.timeout = 30 * time.Second
	}

	return s
}

func (s timeoutStep) execute() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.exe, s.args...)
	cmd.Dir = s.proj
	var errb bytes.Buffer
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", &stepErr{
				step: s.name, msg: "failed time out", cause: context.DeadlineExceeded,
			}
		}

		return "", &stepErr{
			step: s.name, msg: fmt.Sprintf("failed to execute: %s", errb.String()), cause: err,
		}
	}

	return s.message, nil
}
