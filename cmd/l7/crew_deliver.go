package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/crew"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/forge"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/headlessworker"
	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/orchestrationconfig"
	processadapter "github.com/addressanup/level7-dev-loop/internal/l7/adapter/process"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

// discoverForge finds the owner's gh; tests replace it.
var discoverForge = forge.Discover

// pullRequestBase returns the branch checked out at root, which crew pull
// requests target, after checking that pull-request delivery may run.
func pullRequestBase(ctx context.Context, root string, configuration orchestrationconfig.File) (string, error) {
	if !configuration.Features.CrewPR {
		return "", errors.New("pull-request delivery is default OFF; set features.crew_pr to true in .l7/orchestration.json")
	}
	branch, err := crewGitLine(ctx, root, "symbolic-ref", "--short", "-q", "HEAD")
	if err != nil || !domain.CrewBranchValid(branch) {
		return "", errors.New("pull-request delivery needs a checked-out branch to target; HEAD is detached or the branch name is unsafe")
	}
	if _, err := remoteRepository(ctx, root, configuration.EffectiveCrew().Remote); err != nil {
		return "", err
	}
	return branch, nil
}

// remoteRepository names the forge repository behind remote, so gh acts on
// the repository Level 7 pushes to.
func remoteRepository(ctx context.Context, root, remote string) (string, error) {
	url, err := crewGitLine(ctx, root, "remote", "get-url", remote)
	if err != nil {
		return "", fmt.Errorf("remote %s is not configured; set crew.remote or add the remote", remote)
	}
	repository, err := forge.RepositoryFromURL(url)
	if err != nil {
		return "", fmt.Errorf("remote %s: %w", remote, err)
	}
	return repository, nil
}

// checkPullRequestDelivery confirms, before approval, that the flag is on,
// the recorded remote exists, and gh can act for the owner.
func checkPullRequestDelivery(ctx context.Context, root string, configuration orchestrationconfig.File, plan domain.CrewPlan) error {
	if !configuration.Features.CrewPR {
		return errors.New("pull-request delivery is default OFF; set features.crew_pr to true in .l7/orchestration.json")
	}
	repository, err := remoteRepository(ctx, root, plan.Remote)
	if err != nil {
		return err
	}
	client, err := discoverForge(root, repository)
	if err != nil {
		return err
	}
	return client.Authenticated(ctx)
}

func crewMerge(ctx context.Context, location domain.RepositoryLocation, configuration orchestrationconfig.File, store crew.Store, arguments []string) (orchestrationEnvelope, error) {
	if !configuration.Features.CrewPR {
		return orchestrationEnvelope{}, errors.New("pull-request delivery is default OFF; set features.crew_pr to true in .l7/orchestration.json")
	}
	values, confirmed := map[string]string{}, false
	for index := 0; index < len(arguments); index++ {
		switch arguments[index] {
		case "--confirm":
			confirmed = true
		case "--task", "--head":
			if index+1 >= len(arguments) || values[arguments[index]] != "" {
				return orchestrationEnvelope{}, errors.New("crew merge option is missing or duplicated")
			}
			values[arguments[index]] = arguments[index+1]
			index++
		default:
			return orchestrationEnvelope{}, fmt.Errorf("unknown crew merge option %s", arguments[index])
		}
	}
	if !confirmed || values["--task"] == "" || len(values["--head"]) != 40 {
		return orchestrationEnvelope{}, errors.New("crew merge requires --task <id>, the full --head <sha> you reviewed, and --confirm")
	}
	plan, err := activeCrewPlan(store)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	checkpoint, err := store.Checkpoint(plan, values["--task"])
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	executor, err := newCrewMerger(location, configuration)
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	outcome, err := executor.MergePullRequest(ctx, plan, checkpoint, values["--head"])
	if err != nil {
		return orchestrationEnvelope{}, err
	}
	recorded, err := store.Update(plan, checkpoint.TaskID, time.Now(), func(next *domain.CrewCheckpoint) error {
		if !next.State.Terminal() {
			next.State, next.Message, next.Next = domain.CrewDone, outcome.Message, "pull "+plan.TargetBranch+" to get the change"
			next.AttachRequested, next.OwnerEdited = false, false
		}
		return nil
	})
	if err != nil {
		return orchestrationEnvelope{}, fmt.Errorf("pull request merged, but the task state was not recorded: %w", err)
	}
	return passEnvelope("crew merge", "L7-CREW-000", "merged", outcome.Message, "pull "+plan.TargetBranch+" to get the change", recorded), nil
}

// pullRequestMerger is the part of the crew executor that merges.
type pullRequestMerger interface {
	MergePullRequest(context.Context, domain.CrewPlan, domain.CrewCheckpoint, string) (crew.Outcome, error)
}

// newCrewMerger builds the executor that merges; tests replace it.
var newCrewMerger = func(location domain.RepositoryLocation, configuration orchestrationconfig.File) (pullRequestMerger, error) {
	return headlessworker.NewCrew(location.Root, location.CommonDir, configuration)
}

func crewGitLine(ctx context.Context, root string, arguments ...string) (string, error) {
	git, err := processadapter.Resolve("git")
	if err != nil {
		return "", err
	}
	result, err := (processadapter.Runner{}).Run(ctx, processadapter.Request{Executable: git.Path, Arguments: arguments, Directory: root, Environment: processadapter.MinimalEnvironment(), MaxOutputBytes: 64 << 10, Timeout: 30 * time.Second})
	value := strings.TrimSpace(string(result.Stdout))
	if err != nil || result.ExitCode != 0 || value == "" || strings.ContainsAny(value, "\x00\r\n") {
		return "", fmt.Errorf("git %s failed", arguments[0])
	}
	return value, nil
}
