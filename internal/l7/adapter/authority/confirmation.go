package authority

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func (terminal Terminal) ConfirmMerge(ctx context.Context, plan domain.MergePlan) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !safeToken(plan.ChangeID, 64) || !safeToken(plan.TargetRef, 1024) || !strings.HasPrefix(plan.TargetRef, "refs/heads/") || !fullGitID(plan.PreviousCommit) || !fullGitID(plan.Candidate) || plan.PreviousCommit == plan.Candidate {
		return errors.New("merge confirmation plan is invalid")
	}
	if !terminal.interactive {
		return errors.New("merge requires an active terminal interaction")
	}
	if terminal.input == nil || terminal.output == nil {
		return errors.New("merge confirmation terminal is unavailable")
	}
	if _, err := fmt.Fprintf(terminal.output, "Advance local target %s for change %q from %s to %s. Type the full candidate SHA to continue: ", plan.TargetRef, plan.ChangeID, plan.PreviousCommit, plan.Candidate); err != nil {
		return errors.New("cannot write merge confirmation prompt")
	}
	reader := bufio.NewReader(io.LimitReader(terminal.input, 1024))
	answer, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.New("cannot read merge confirmation response")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(answer) != plan.Candidate {
		return errors.New("full candidate SHA was not confirmed")
	}
	return nil
}
