package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func (adapter Adapter) InspectMerge(ctx context.Context, request domain.MergeRequest) (domain.MergeTarget, error) {
	if err := validateMergeRequest(request); err != nil {
		return domain.MergeTarget{}, err
	}
	limited := adapter.WithLimits(request.MaxOutputBytes, adapter.maxPaths)
	normalized, err := limited.singleLine(ctx, request.Root, "check-ref-format", "--branch", request.TargetBranch)
	if err != nil || normalized != request.TargetBranch {
		return domain.MergeTarget{}, errors.New("merge target is not one exact local branch name")
	}
	targetRef := "refs/heads/" + normalized
	current, err := limited.singleLine(ctx, request.Root, "show-ref", "--verify", "--hash", targetRef)
	if err != nil || !fullObjectID(current) {
		return domain.MergeTarget{}, errors.New("merge target local branch does not exist")
	}
	for label, identity := range map[string]string{"expected old": request.ExpectedOld, "candidate": request.Candidate} {
		resolved, resolveErr := limited.singleLine(ctx, request.Root, "rev-parse", "--verify", identity+"^{commit}")
		if resolveErr != nil || resolved != identity {
			return domain.MergeTarget{}, fmt.Errorf("merge %s commit is unavailable", label)
		}
	}
	ancestor, err := limited.isAncestor(ctx, request.Root, request.ExpectedOld, request.Candidate)
	if err != nil || !ancestor {
		return domain.MergeTarget{}, errors.New("merge candidate is not a fast-forward descendant of the expected target")
	}
	if current != request.ExpectedOld && current != request.Candidate {
		return domain.MergeTarget{}, errors.New("merge target diverged from the expected old commit")
	}
	worktrees, err := limited.run(ctx, request.Root, "worktree", "list", "--porcelain")
	if err != nil {
		return domain.MergeTarget{}, errors.New("cannot inspect checked-out worktree branches")
	}
	checkedOut, err := refCheckedOut(worktrees, targetRef)
	if err != nil {
		return domain.MergeTarget{}, err
	}
	if checkedOut {
		return domain.MergeTarget{}, errors.New("merge target is checked out in a worktree")
	}
	return domain.MergeTarget{
		Branch: normalized, Ref: targetRef, CurrentCommit: current, ExpectedOld: request.ExpectedOld,
		Candidate: request.Candidate, AlreadyAdvanced: current == request.Candidate,
	}, nil
}

func (adapter Adapter) AdvanceMerge(ctx context.Context, request domain.MergeRequest) error {
	target, err := adapter.InspectMerge(ctx, request)
	if err != nil {
		return err
	}
	if target.AlreadyAdvanced || target.CurrentCommit != request.ExpectedOld {
		return errors.New("merge target is not at the expected old commit")
	}
	limited := adapter.WithLimits(request.MaxOutputBytes, adapter.maxPaths)
	if _, err := limited.run(ctx, request.Root, "update-ref", "--no-deref", target.Ref, request.Candidate, request.ExpectedOld); err != nil {
		return fmt.Errorf("atomic merge ref update failed: %w", err)
	}
	current, err := limited.singleLine(ctx, request.Root, "show-ref", "--verify", "--hash", target.Ref)
	if err != nil || current != request.Candidate {
		return errors.New("merge ref update completed without the expected candidate identity")
	}
	return nil
}

func (adapter Adapter) MergeCurrent(ctx context.Context, root string, receipt domain.MergeReceipt, maxOutput int) (bool, error) {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || !domain.MergeReceiptValid(receipt) || maxOutput < 64<<10 || maxOutput > 64<<20 {
		return false, errors.New("merge receipt query is invalid")
	}
	limited := adapter.WithLimits(maxOutput, adapter.maxPaths)
	current, err := limited.singleLine(ctx, root, "show-ref", "--verify", "--hash", receipt.TargetRef)
	if err != nil {
		return false, nil
	}
	tree, err := limited.CommitTree(ctx, root, receipt.Candidate.Commit)
	if err != nil {
		return false, err
	}
	return current == receipt.Candidate.Commit && tree == receipt.Candidate.Tree, nil
}

func validateMergeRequest(request domain.MergeRequest) error {
	if !filepath.IsAbs(request.Root) || filepath.Clean(request.Root) != request.Root || !fullObjectID(request.ExpectedOld) || !fullObjectID(request.Candidate) || request.ExpectedOld == request.Candidate || request.MaxOutputBytes < 64<<10 || request.MaxOutputBytes > 64<<20 || len(request.TargetBranch) < 1 || len(request.TargetBranch) > 1000 || strings.HasPrefix(request.TargetBranch, "-") || strings.HasPrefix(request.TargetBranch, "refs/") || strings.TrimSpace(request.TargetBranch) != request.TargetBranch || strings.ContainsAny(request.TargetBranch, "\r\n\x00") || !utf8.ValidString(request.TargetBranch) {
		return errors.New("controlled merge request is invalid")
	}
	return nil
}

func refCheckedOut(data []byte, targetRef string) (bool, error) {
	if len(data) == 0 || !utf8.Valid(data) || !strings.HasSuffix(string(data), "\n") {
		return false, errors.New("Git returned an invalid worktree listing")
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		if !strings.HasPrefix(line, "branch ") {
			continue
		}
		value := strings.TrimPrefix(line, "branch ")
		if value == "" || strings.ContainsAny(value, "\r\n\x00") {
			return false, errors.New("Git returned an invalid worktree branch")
		}
		if value == targetRef {
			return true, nil
		}
	}
	return false, nil
}
