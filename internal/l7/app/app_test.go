package app

import (
	"context"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestExecuteWithoutLifecyclePortsRemainsTruthfullyInert(t *testing.T) {
	application := New("test-version")
	tests := []struct {
		command domain.Command
		outcome domain.Outcome
		code    string
		state   string
		exit    int
	}{
		{domain.CommandHelp, domain.OutcomePass, "L7-CLI-000", "available", 0},
		{domain.CommandVersion, domain.OutcomePass, "L7-CLI-000", "available", 0},
		{domain.CommandAdopt, domain.OutcomeBlocked, "L7-CAP-001", "unavailable", 2},
		{domain.CommandBrief, domain.OutcomeBlocked, "L7-CAP-001", "unavailable", 2},
		{domain.CommandStatus, domain.OutcomeBlocked, "L7-STATUS-001", "unavailable", 2},
		{domain.Command("run"), domain.OutcomeFailed, "L7-CLI-001", "invalid", 1},
	}
	for _, test := range tests {
		result := application.Execute(context.Background(), test.command)
		if result.Schema != domain.ResultSchema || result.Outcome != test.outcome || result.Code != test.code || result.State != test.state || result.Version != "test-version" || result.ExitCode() != test.exit {
			t.Fatalf("command %q: unexpected result: %+v", test.command, result)
		}
	}
}

func TestExecuteStopsBeforeWorkWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := New("test-version").Execute(ctx, domain.CommandStatus)
	if result.Outcome != domain.OutcomeCancelled || result.Code != "L7-CLI-003" || result.ExitCode() != 130 {
		t.Fatalf("cancelled execution advanced: %+v", result)
	}
}

func TestNewUsesDeterministicDevelopmentVersion(t *testing.T) {
	result := New("  ").Execute(context.Background(), domain.CommandVersion)
	if result.Version != developmentVersion {
		t.Fatalf("version=%q, want %q", result.Version, developmentVersion)
	}
}

func TestInvalidPreservesBoundedCommandContext(t *testing.T) {
	command := "--" + string(make([]byte, maxCommandBytes*2))
	result := New("test-version").Invalid(command, "unknown flag")
	if result.Outcome != domain.OutcomeFailed || len(result.Command) != maxCommandBytes || result.Message != "unknown flag" || result.Next != "run l7 help" {
		t.Fatalf("unexpected invalid result: %+v", result)
	}
}
