package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var pushTimeout = 10 * time.Second

func main() {
	if err := checkOS(); err != nil {
		log.Fatalf("An error: %v", err)
	}

	proj := flag.String("p", "", "Project directory")
	flag.Parse()

	if err := run(*proj, os.Stdout); err != nil {
		log.Fatalf("An error: %v", err)
	}
}

func run(proj string, out io.Writer) error {
	if proj == "" {
		return fmt.Errorf("The project directory is required: %w", ErrValidation)
	}

	pipeline := make([]executer, 4)

	pipeline[0] = newStep(
		"go build", "go", "Go Build: SUCCESS", proj, []string{"build", ".", "errors"},
	)
	pipeline[1] = newStep(
		"go test", "go", "Go Test: SUCCESS", proj, []string{"test", "-v"},
	)
	pipeline[2] = newExceptionStep(
		"go fmt", "gofmt", "Gofmt: SUCCESS", proj, []string{"-l", "."},
	)
	pipeline[3] = newTimeoutStep(
		"git push", "git", "Git Push: SUCCESS", proj,
		[]string{"push", "origin", "master"}, pushTimeout,
	)

	sig := make(chan os.Signal, 1)
	errCh := make(chan error)
	done := make(chan struct{})

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for _, s := range pipeline {
			msg, err := s.execute()
			if err != nil {
				errCh <- err
				return
			}

			_, err = fmt.Fprintln(out, msg)
			if err != nil {
				errCh <- err
				return
			}
		}
		close(done)
	}()

	for {
		select {
		case rec := <-sig:
			signal.Stop(sig)
			return fmt.Errorf("%s: Exiting: %w", rec, ErrSignal)
		case err := <-errCh:
			return err
		case <-done:
			return nil
		}
	}
}

func checkOS() error {
	if runtime.GOOS == "windows" {
		return ErrUnsupportedOs
	}

	return nil
}

type executer interface {
	execute() (string, error)
}
