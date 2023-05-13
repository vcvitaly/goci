package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

type step struct {
	name    string
	exe     string
	args    []string
	message string
	proj    string
}

func newStep(name, exe, message, proj string, args []string) step {
	return step{
		name:    name,
		exe:     exe,
		message: message,
		proj:    proj,
		args:    args,
	}
}

func (s step) execute() (string, error) {
	cmd := exec.Command("go", s.args...)
	cmd.Dir = s.proj
	var errb bytes.Buffer
	cmd.Stderr = &errb

	if err := cmd.Run(); err != nil {
		return "", &stepErr{step: s.name, msg: fmt.Sprintf("failed to execute: %s", errb.String()), cause: err}
	}

	return s.message, nil
}
