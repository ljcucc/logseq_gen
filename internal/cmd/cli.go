package cmd

import (
	"fmt"
	"strings"
)

// Runner is the interface for the command runner.
type Runner interface {
	Build() error
	Clear() error
}

// Run executes the command-line interface.
func Run(g Runner, args []string) error {
	command := "build"
	if len(args) > 1 {
		command = strings.ToLower(args[1])
	}

	switch command {
	case "build":
		return g.Build()
	case "clear":
		return g.Clear()
	default:
		return fmt.Errorf("unknown command: %s\nUsage: %s [build|clear]", command, args[0])
	}
}
