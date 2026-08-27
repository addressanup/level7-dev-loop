package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	cliapp "github.com/addressanup/level7-dev-loop/internal/l7/app"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
	"github.com/addressanup/level7-dev-loop/internal/l7/presentation"
)

var version = "0.1.0-dev"

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, arguments []string, stdout, stderr io.Writer) int {
	application := cliapp.New(version)
	command, jsonOutput, invalid := parseArguments(arguments, application)
	result := application.Execute(ctx, command)
	if invalid != nil {
		result = *invalid
	}

	var output []byte
	if jsonOutput {
		var err error
		output, err = presentation.JSON(result)
		if err != nil {
			fmt.Fprintf(stderr, "FAILED code=L7-CLI-002 message=%q\n", "cannot render JSON output")
			return 1
		}
	} else {
		output = presentation.Text(result)
	}
	if _, err := stdout.Write(output); err != nil {
		fmt.Fprintf(stderr, "FAILED code=L7-CLI-002 message=%q\n", "cannot write command output")
		return 1
	}
	return result.ExitCode()
}

func parseArguments(arguments []string, application cliapp.Application) (domain.Command, bool, *domain.Result) {
	var positional []string
	jsonOutput := false
	for _, argument := range arguments {
		switch argument {
		case "--json":
			if jsonOutput {
				result := application.Invalid(argument, "duplicate --json flag")
				return "", true, &result
			}
			jsonOutput = true
		case "--help", "-h":
			positional = append(positional, string(domain.CommandHelp))
		case "--version":
			positional = append(positional, string(domain.CommandVersion))
		default:
			if strings.HasPrefix(argument, "-") {
				result := application.Invalid(argument, "unknown flag")
				return "", jsonOutput, &result
			}
			positional = append(positional, argument)
		}
	}

	if len(positional) == 0 {
		return domain.CommandHelp, jsonOutput, nil
	}
	if len(positional) > 1 {
		result := application.Invalid(strings.Join(positional, " "), "expected exactly one command")
		return "", jsonOutput, &result
	}
	return domain.Command(positional[0]), jsonOutput, nil
}
