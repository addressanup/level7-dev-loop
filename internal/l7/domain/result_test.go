package domain

import "testing"

func TestResultExitCode(t *testing.T) {
	tests := []struct {
		outcome Outcome
		want    int
	}{
		{OutcomePass, 0},
		{OutcomeBlocked, 2},
		{OutcomeFailed, 1},
		{OutcomeCancelled, 130},
		{Outcome("UNKNOWN"), 1},
	}
	for _, test := range tests {
		result := Result{Outcome: test.outcome}
		if got := result.ExitCode(); got != test.want {
			t.Fatalf("outcome %q: ExitCode()=%d, want %d", test.outcome, got, test.want)
		}
	}
}

func TestResultSchemaChangesWhenWaveFourAddsReadinessDetails(t *testing.T) {
	if ResultSchema != 4 {
		t.Fatalf("ResultSchema=%d, want 4", ResultSchema)
	}
}

func TestExecutionCommandsAreExplicit(t *testing.T) {
	for _, command := range []Command{CommandRun, CommandVerify, CommandReview, CommandReady, CommandMerge} {
		if command == "" {
			t.Fatal("execution command is empty")
		}
	}
}
