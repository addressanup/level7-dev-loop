package verify

import (
	"context"
	"errors"
	"strings"
	"testing"

	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestRunnerUsesExactArgvAndStopsAtFirstFailure(t *testing.T) {
	var requests []processadapter.Request
	runner := New(
		func(name string) (processadapter.Executable, error) {
			return processadapter.Executable{Path: "/usr/bin/" + name, Digest: strings.Repeat("a", 64)}, nil
		},
		func(_ context.Context, request processadapter.Request) (processadapter.Result, error) {
			requests = append(requests, request)
			if len(requests) == 2 {
				return processadapter.Result{ExitCode: 4, Stderr: []byte("failed\nsecret detail")}, nil
			}
			return processadapter.Result{ExitCode: 0}, nil
		},
	)
	commands := []domain.VerificationCommand{
		{Name: "lint", Argv: []string{"make", "lint"}},
		{Name: "test", Argv: []string{"make", "test"}},
		{Name: "build", Argv: []string{"make", "build"}},
	}
	checks, err := runner.Run(context.Background(), "/repo", commands, 1<<20, 30)
	if err == nil || len(checks) != 2 || len(requests) != 2 || requests[0].Executable != "/usr/bin/make" || strings.Join(requests[0].Arguments, "|") != "lint" || !checks[0].Passed || checks[1].Passed || checks[1].ExitCode != 4 {
		t.Fatalf("Run() checks=%+v requests=%+v error=%v", checks, requests, err)
	}
}

func TestRunnerFailsClosedOnResolutionAndCancellation(t *testing.T) {
	runner := New(func(string) (processadapter.Executable, error) {
		return processadapter.Executable{}, errors.New("missing")
	}, func(context.Context, processadapter.Request) (processadapter.Result, error) {
		t.Fatal("run called after failed resolution")
		return processadapter.Result{}, nil
	})
	checks, err := runner.Run(context.Background(), "/repo", []domain.VerificationCommand{{Name: "test", Argv: []string{"missing"}}}, 1<<20, 30)
	if err == nil || len(checks) != 1 || checks[0].Code != "L7-VERIFY-002" {
		t.Fatalf("resolution checks=%+v error=%v", checks, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runner = New(func(name string) (processadapter.Executable, error) {
		return processadapter.Executable{Path: "/usr/bin/" + name}, nil
	}, func(context.Context, processadapter.Request) (processadapter.Result, error) {
		t.Fatal("run called after cancellation")
		return processadapter.Result{}, nil
	})
	if checks, err := runner.Run(ctx, "/repo", []domain.VerificationCommand{{Name: "test", Argv: []string{"make"}}}, 1<<20, 30); !errors.Is(err, context.Canceled) || len(checks) != 0 {
		t.Fatalf("cancelled checks=%+v error=%v", checks, err)
	}
}

func BenchmarkVerificationDispatch(b *testing.B) {
	runner := New(func(name string) (processadapter.Executable, error) {
		return processadapter.Executable{Path: "/usr/bin/" + name}, nil
	}, func(context.Context, processadapter.Request) (processadapter.Result, error) {
		return processadapter.Result{ExitCode: 0}, nil
	})
	commands := []domain.VerificationCommand{{Name: "test", Argv: []string{"make", "test"}}}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := runner.Run(context.Background(), "/repo", commands, 1<<20, 30); err != nil {
			b.Fatal(err)
		}
	}
}
