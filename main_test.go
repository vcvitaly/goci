package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skipf("Git is not installed. Skipping test.")
	}

	var testCases = []struct {
		name     string
		proj     string
		out      string
		expErr   error
		setupGit bool
	}{
		{name: "success", proj: "./testdata/tool/",
			out:    "Go Build: SUCCESS\nGo Test: SUCCESS\nGofmt: SUCCESS\nGit Push: SUCCESS\n",
			expErr: nil, setupGit: true},
		{name: "fail", proj: "./testdata/toolErr/", out: "", expErr: &stepErr{step: "go build"}, setupGit: false},
		{name: "failFormat", proj: "./testdata/toolFmtErr/", out: "", expErr: &stepErr{step: "go fmt"}, setupGit: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := tc.proj
			if tc.setupGit {
				var cleanup func() error
				runDir, cleanup = setupGit(t, tc.proj)
				defer func() {
					err = cleanup()
					if err != nil {
						t.Fatal(err)
					}
				}()
			}

			var out bytes.Buffer
			err := run(runDir, &out)

			if tc.expErr != nil {
				if err == nil {
					t.Errorf("Expected error: %q. Got 'nil' instead.", tc.expErr)
					return
				}

				if !errors.Is(err, tc.expErr) {
					t.Errorf("Expected error: %q. Got %q.", tc.expErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %q", err)
			}

			if out.String() != tc.out {
				t.Errorf("Expected output: %q. Got %q", tc.out, out.String())
			}
		})
	}
}

func setupGit(t *testing.T, proj string) (string, func() error) {
	t.Helper()

	gitExec, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}

	tempRemoteDir, err := os.MkdirTemp("", "gocitest_remote")
	if err != nil {
		t.Fatal(err)
	}

	tempLocalDir, err := os.MkdirTemp("", "gocitest_local")
	if err != nil {
		t.Fatal(err)
	}

	projPath, err := filepath.Abs(proj)
	if err != nil {
		t.Fatal(err)
	}

	err = copyDir(projPath, tempLocalDir)
	if err != nil {
		t.Fatal(err)
	}

	remoteURI := fmt.Sprintf("file://%s", tempRemoteDir)

	var gitCmdList = []struct {
		args []string
		dir  string
		env  []string
	}{
		{[]string{"init", "--bare"}, tempRemoteDir, nil},
		{[]string{"init"}, tempLocalDir, nil},
		{[]string{"remote", "add", "origin", remoteURI}, tempLocalDir, nil},
		{[]string{"add", "."}, tempLocalDir, nil},
		{[]string{"commit", "-m", "test"}, tempLocalDir,
			[]string{
				"GIT_COMMITTER_NAME=test",
				"GIT_COMMITTER_EMAIL=test@example.com",
				"GIT_AUTHOR_NAME=test",
				"GIT_AUTHOR_EMAIL=test@example.com",
			}},
	}

	for _, g := range gitCmdList {
		gitCmd := exec.Command(gitExec, g.args...)
		gitCmd.Dir = g.dir
		var errb bytes.Buffer
		gitCmd.Stderr = &errb

		if g.env != nil {
			gitCmd.Env = append(os.Environ(), g.env...)
		}

		err := gitCmd.Run()
		if err != nil {
			t.Fatal(fmt.Errorf("An error while running %s:\nstderr: %s\nerr:%s\n", gitExec, errb.String(), err))
		}
	}

	return tempLocalDir, func() error {
		err := os.RemoveAll(tempRemoteDir)
		if err != nil {
			return err
		}
		err = os.RemoveAll(tempLocalDir)
		if err != nil {
			return err
		}
		return nil
	}
}

func copyDir(from string, to string) error {
	cpExec, err := exec.LookPath("cp")
	if err != nil {
		return err
	}

	cpCmd := exec.Command(cpExec, "-r", from+"/", to+"/")
	if err := cpCmd.Run(); err != nil {
		return err
	}

	return nil
}
