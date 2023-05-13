package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
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

	pipeline := make([]executer, 3)

	pipeline[0] = newStep(
		"go build", "go", "Go Build: SUCCESS", proj, []string{"build", ".", "errors"},
	)
	pipeline[1] = newStep(
		"go test", "go", "Go Test: SUCCESS", proj, []string{"test", "-v"},
	)
	pipeline[2] = newExceptionStep(
		"go fmt", "gofmt", "Gofmt: SUCCESS", proj, []string{"-l", "."},
	)

	for _, s := range pipeline {
		msg, err := s.execute()
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(out, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

type executer interface {
	execute() (string, error)
}
