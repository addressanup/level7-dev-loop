package authority

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestMergeConfirmationRequiresInteractiveFullCandidateSHA(t *testing.T) {
	plan := mergePlan()
	var output bytes.Buffer
	terminal := NewTerminal(strings.NewReader(plan.Candidate+"\n"), &output, true, "accountable-owner")
	if err := terminal.ConfirmMerge(context.Background(), plan); err != nil || !strings.Contains(output.String(), plan.TargetRef) || !strings.Contains(output.String(), plan.PreviousCommit) || !strings.Contains(output.String(), plan.Candidate) {
		t.Fatalf("ConfirmMerge() output=%q error=%v", output.String(), err)
	}
	for _, test := range []Terminal{
		NewTerminal(strings.NewReader("short\n"), &output, true, "accountable-owner"),
		NewTerminal(strings.NewReader(plan.Candidate+"\n"), &output, false, "accountable-owner"),
		NewTerminal(nil, &output, true, "accountable-owner"),
	} {
		if err := test.ConfirmMerge(context.Background(), plan); err == nil {
			t.Fatal("ConfirmMerge unexpectedly accepted invalid authority")
		}
	}
}

func TestMergeConfirmationHonorsCancellationBeforePrompt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var output bytes.Buffer
	if err := NewTerminal(strings.NewReader(mergePlan().Candidate+"\n"), &output, true, "accountable-owner").ConfirmMerge(ctx, mergePlan()); err != context.Canceled || output.Len() != 0 {
		t.Fatalf("ConfirmMerge() output=%q error=%v", output.String(), err)
	}
}

func mergePlan() domain.MergePlan {
	return domain.MergePlan{ChangeID: "product-change", TargetRef: "refs/heads/main", PreviousCommit: strings.Repeat("a", 40), Candidate: strings.Repeat("b", 40)}
}
