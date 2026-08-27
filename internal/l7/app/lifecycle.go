package app

import (
	"context"
	"strings"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

type Ports struct {
	Locate             func(context.Context, string) (domain.RepositoryLocation, error)
	Snapshot           func(context.Context, string, string, int, int) (domain.RepositorySnapshot, error)
	LoadConfiguration  func(string) (domain.Configuration, bool, error)
	AdoptConfiguration func(string, bool) (domain.Configuration, bool, error)
	LoadActive         func(string) (domain.ActiveChange, bool, error)
	SaveActive         func(string, domain.ActiveChange) error
	Acquire            func(string) (func() error, error)
	EnsureBrief        func(string, domain.ChangeBrief) (bool, error)
	LoadBrief          func(string, string) (domain.ChangeBrief, error)
	Pending            func(context.Context, string, int, int) (domain.PendingChanges, error)
	PathCommit         func(context.Context, string, string) (string, error)
	Commit             func(context.Context, domain.CommitRequest) (domain.RepositoryLocation, error)
	CommitMatches      func(context.Context, string, string, string, string, int, int) (bool, error)
	CommitPaths        func(context.Context, string, string, string, int, int) ([]string, error)
	CommitTree         func(context.Context, string, string) (string, error)
	PathSetDigest      func([]string) string
	ConfirmApproval    func(context.Context, string, domain.Provider, string) (domain.ApprovalBinding, error)
	LoadApproval       func(string) (domain.ApprovalBinding, bool, error)
	SaveApproval       func(string, domain.ApprovalBinding) error
	LoadRun            func(string) (domain.RunEvidence, bool, error)
	SaveRun            func(string, domain.RunEvidence) error
	LoadVerification   func(string) (domain.VerificationEvidence, bool, error)
	SaveVerification   func(string, domain.VerificationEvidence) error
	LoadReview         func(string) (domain.ReviewEvidence, bool, error)
	SaveReview         func(string, domain.ReviewEvidence) error
	RunProvider        func(context.Context, domain.ProviderTask, int, int) (domain.ProviderResponse, error)
	RunVerification    func(context.Context, string, []domain.VerificationCommand, int, int) ([]domain.CheckResult, error)
	WriteVerification  func(string, domain.VerificationEvidence, string) (string, error)
	WriteAudit         func(string, domain.ReviewEvidence) (string, error)
}

var builtinProtectedPaths = []string{
	".codex/**",
	".claude/**",
	".claude-plugin/plugin.json",
	".codex-plugin/plugin.json",
	".git/**",
	".gitattributes",
	".github/CODEOWNERS",
	".github/workflows/**",
	".gitignore",
	".gitmodules",
	".l7/**",
	"AGENTS.md",
	"CLAUDE.md",
	"Makefile",
	"deploy/**",
	"harness/**",
	"internal/auth/**",
	"internal/authorization/**",
	"internal/harness/**",
	"internal/security/**",
	"marketplace.json",
	"migrations/**",
	"plugin.json",
	"scripts/harness/**",
	"skills/**",
}

func (application Application) adopt(ctx context.Context, request domain.Request) domain.Result {
	if !application.lifecyclePortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-CAP-001", string(request.Command), "unavailable", "repository adoption is not configured in this build", "run l7 help")
	}
	location, err := application.ports.Locate(ctx, application.cwd)
	if err != nil {
		return application.failure("L7-REPO-001", request.Command, "unavailable", "cannot adopt repository: "+bounded(err.Error(), 512), "run l7 adopt from an existing non-bare Git worktree with an initial commit")
	}
	release, err := application.ports.Acquire(location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait for the active mutation to finish, then run l7 adopt again")
	}
	released := false
	defer func() {
		if !released {
			_ = release()
		}
	}()
	rechecked, err := application.ports.Locate(ctx, application.cwd)
	if err != nil || !sameLocation(location, rechecked) {
		return application.failure("L7-GIT-002", request.Command, "changed", "Git identity changed before adoption", "retry l7 adopt against a stable repository")
	}
	configuration, changed, err := application.ports.AdoptConfiguration(location.Root, request.EnableLocalLifecycle)
	if err != nil {
		return application.failure("L7-CONFIG-001", request.Command, "invalid", "repository configuration was not changed: "+bounded(err.Error(), 512), "repair the conflicting .l7/config.json, then retry l7 adopt")
	}
	storedConfiguration, found, loadErr := application.ports.LoadConfiguration(location.Root)
	finalLocation, locateErr := application.ports.Locate(ctx, application.cwd)
	if loadErr != nil || !found || locateErr != nil || !sameConfiguration(configuration, storedConfiguration) || !sameLocation(location, finalLocation) {
		return application.failure("L7-CONFIG-003", request.Command, "recovery-required", "repository or configuration changed during adoption", "inspect .l7/config.json and run l7 status before continuing")
	}
	if err := release(); err != nil {
		released = true
		return application.failure("L7-STATE-003", request.Command, "adopted", "repository was adopted but lock release reported: "+bounded(err.Error(), 512), "run l7 status before continuing")
	}
	released = true
	state := "adopted"
	message := "repository configuration is valid"
	if changed {
		message = "repository configuration was written atomically"
	}
	if configuration.LocalLifecycle {
		state = "enabled"
	}
	result := application.result(domain.OutcomePass, "L7-ADOPT-000", string(request.Command), state, message, "run l7 status")
	result.Repository = repositoryDetails(location)
	return result
}

func (application Application) createBrief(ctx context.Context, request domain.Request) domain.Result {
	if !application.lifecyclePortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-CAP-001", string(request.Command), "unavailable", "change intake is not configured in this build", "run l7 help")
	}
	if !request.Tier.Valid() || request.ChangeID == "" || request.Problem == "" || len(request.Scope) == 0 {
		return application.Invalid(string(request.Command), "brief requires a change ID, risk tier, problem, and scope")
	}
	if request.Tier == domain.TierRoutine && (len(request.AcceptanceCriteria) != 0 || len(request.Risks) != 0 || len(request.Rollback) != 0) {
		return application.Invalid(string(request.Command), "Tier 1 stores only a concise task, base, and scope; omit artifact-only fields")
	}
	location, err := application.ports.Locate(ctx, application.cwd)
	if err != nil {
		return application.failure("L7-REPO-001", request.Command, "unavailable", "cannot inspect repository: "+bounded(err.Error(), 512), "run l7 brief from an adopted Git worktree")
	}
	configuration, found, err := application.ports.LoadConfiguration(location.Root)
	if err != nil {
		return application.failure("L7-CONFIG-001", request.Command, "invalid", "repository configuration is invalid: "+bounded(err.Error(), 512), "repair .l7/config.json, then retry l7 brief")
	}
	if !found {
		return application.result(domain.OutcomeBlocked, "L7-CONFIG-002", string(request.Command), "unadopted", "repository has not been adopted", "run l7 adopt --enable-local-lifecycle")
	}
	if !configuration.LocalLifecycle {
		return application.result(domain.OutcomeBlocked, "L7-FLAG-001", string(request.Command), "disabled", "local lifecycle behavior is default OFF", "run l7 adopt --enable-local-lifecycle")
	}
	if requestBytes(request) > configuration.MaxInputBytes {
		return application.result(domain.OutcomeBlocked, "L7-INPUT-001", string(request.Command), "bounded", "change input exceeds the repository's configured size limit", "shorten the change input or explicitly raise the bounded configuration limit")
	}
	if request.Tier != domain.TierHighRisk && touchesProtected(request.Scope, configuration.ProtectedPaths) {
		return application.result(domain.OutcomeBlocked, "L7-RISK-001", string(request.Command), "risk-mismatch", "declared scope intersects a protected control", "declare Tier 3 and obtain explicit owner approval before implementation")
	}
	expectedBriefPath := ""
	if request.Tier != domain.TierRoutine {
		expectedBriefPath = "docs/artifacts/changes/" + request.ChangeID + ".md"
	}
	snapshot, err := application.ports.Snapshot(ctx, location.Root, location.Head, configuration.MaxGitOutputBytes, configuration.MaxGitPaths)
	if err != nil {
		return application.failure("L7-GIT-001", request.Command, "invalid", "cannot establish a clean brief base: "+bounded(err.Error(), 512), "stabilize the Git repository, then retry l7 brief")
	}
	if !onlyPermitted(snapshot.ChangedPaths, expectedBriefPath) {
		result := application.result(domain.OutcomeBlocked, "L7-BRIEF-001", string(request.Command), "dirty", "change intake requires a clean worktree", "commit or restore current changes, then run l7 brief again")
		result.Repository = detailsForSnapshot(snapshot, "", request.Tier, request.Scope, nil)
		return result
	}

	release, err := application.ports.Acquire(location.CommonDir)
	if err != nil {
		return application.result(domain.OutcomeBlocked, "L7-STATE-002", string(request.Command), "busy", "another Level 7 mutation is active or the repository lock is unsafe", "wait for the active mutation to finish, then run l7 brief again")
	}
	released := false
	defer func() {
		if !released {
			_ = release()
		}
	}()
	recheckedLocation, err := application.ports.Locate(ctx, application.cwd)
	if err != nil || !sameLocation(location, recheckedLocation) {
		return application.failure("L7-GIT-002", request.Command, "changed", "Git identity changed before change intake", "retry l7 brief against a stable repository")
	}
	recheckedConfiguration, found, err := application.ports.LoadConfiguration(location.Root)
	if err != nil || !found || !sameConfiguration(configuration, recheckedConfiguration) {
		return application.failure("L7-CONFIG-003", request.Command, "changed", "repository configuration changed during change intake", "retry l7 brief against stable configuration")
	}
	recheckedSnapshot, err := application.ports.Snapshot(ctx, location.Root, location.Head, configuration.MaxGitOutputBytes, configuration.MaxGitPaths)
	if err != nil || !sameSnapshot(snapshot, recheckedSnapshot) || !onlyPermitted(recheckedSnapshot.ChangedPaths, expectedBriefPath) {
		return application.failure("L7-GIT-002", request.Command, "changed", "repository scope changed during change intake", "restore or commit concurrent changes, then retry l7 brief")
	}

	active := domain.ActiveChange{Kind: domain.ActiveTierOne, ID: request.ChangeID, Tier: request.Tier, Base: location.Head, Problem: request.Problem, Scope: append([]string{}, request.Scope...)}
	briefPath := ""
	plannedBrief := domain.ChangeBrief{}
	if request.Tier != domain.TierRoutine {
		briefPath = expectedBriefPath
		plannedBrief = domain.ChangeBrief{
			ID:                 request.ChangeID,
			Tier:               request.Tier,
			Base:               location.Head,
			Path:               briefPath,
			Problem:            request.Problem,
			Scope:              append([]string{}, request.Scope...),
			AcceptanceCriteria: append([]string{}, request.AcceptanceCriteria...),
			Risks:              append([]string{}, request.Risks...),
			Rollback:           append([]string{}, request.Rollback...),
		}
		if _, err := application.ports.EnsureBrief(location.Root, plannedBrief); err != nil {
			return application.failure("L7-BRIEF-002", request.Command, "invalid", "change brief was not created: "+bounded(err.Error(), 512), "correct the brief input or resolve the existing brief collision")
		}
		active = domain.ActiveChange{Kind: domain.ActiveBrief, ID: request.ChangeID, BriefPath: briefPath}
	}
	if err := application.ports.SaveActive(location.CommonDir, active); err != nil {
		return application.failure("L7-STATE-001", request.Command, "recovery-required", "active context was not recorded: "+bounded(err.Error(), 512), "rerun the identical l7 brief command to recover without overwriting the brief")
	}
	finalSnapshot, err := application.ports.Snapshot(ctx, location.Root, location.Head, configuration.MaxGitOutputBytes, configuration.MaxGitPaths)
	if err != nil || !sameCandidate(location, finalSnapshot.RepositoryLocation) || !onlyPermitted(finalSnapshot.ChangedPaths, briefPath) {
		return application.failure("L7-GIT-002", request.Command, "recovery-required", "repository changed while recording the brief", "inspect l7 status and restore unauthorized concurrent changes before continuing")
	}
	storedActive, activeFound, activeErr := application.ports.LoadActive(location.CommonDir)
	if activeErr != nil || !activeFound || !sameStoredActive(active, storedActive) {
		return application.failure("L7-STATE-004", request.Command, "recovery-required", "active context changed while recording the brief", "inspect .git/l7/product/active.json before continuing")
	}
	if plannedBrief.ID != "" {
		storedBrief, briefErr := application.ports.LoadBrief(location.Root, briefPath)
		if briefErr != nil || !sameBrief(plannedBrief, storedBrief) {
			return application.failure("L7-BRIEF-004", request.Command, "recovery-required", "change brief changed while recording active context", "restore the exact requested brief before continuing")
		}
	}
	if err := release(); err != nil {
		released = true
		return application.failure("L7-STATE-003", request.Command, "planned", "change was recorded but lock release reported: "+bounded(err.Error(), 512), "run l7 status before implementation")
	}
	released = true
	next, _ := domain.NextTransition(request.Tier, domain.StatePlanned)
	result := application.result(domain.OutcomePass, "L7-BRIEF-000", string(request.Command), string(domain.StatePlanned), "change context recorded", next.Action)
	result.Repository = detailsForSnapshot(finalSnapshot, request.ChangeID, request.Tier, request.Scope, nil)
	return result
}

func (application Application) status(ctx context.Context, request domain.Request) domain.Result {
	if !application.lifecyclePortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-STATUS-001", string(request.Command), "unavailable", "repository workflow status is not configured in this build", "run l7 help")
	}
	location, err := application.ports.Locate(ctx, application.cwd)
	if err != nil {
		return application.failure("L7-REPO-001", request.Command, "unavailable", "cannot inspect repository: "+bounded(err.Error(), 512), "run l7 status from an adopted Git worktree")
	}
	configuration, found, err := application.ports.LoadConfiguration(location.Root)
	if err != nil {
		return application.failure("L7-CONFIG-001", request.Command, "invalid", "repository configuration is invalid: "+bounded(err.Error(), 512), "repair .l7/config.json, then run l7 status")
	}
	if !found {
		result := application.result(domain.OutcomeBlocked, "L7-CONFIG-002", string(request.Command), "unadopted", "repository has not been adopted", "run l7 adopt --enable-local-lifecycle")
		result.Repository = repositoryDetails(location)
		return result
	}
	if !configuration.LocalLifecycle {
		result := application.result(domain.OutcomeBlocked, "L7-FLAG-001", string(request.Command), "disabled", "local lifecycle behavior is default OFF", "run l7 adopt --enable-local-lifecycle")
		result.Repository = repositoryDetails(location)
		return result
	}
	initialActive, activeFound, err := application.ports.LoadActive(location.CommonDir)
	if err != nil {
		return application.failure("L7-STATE-001", request.Command, "invalid", "active context is invalid: "+bounded(err.Error(), 512), "repair or explicitly recover .git/l7/product/active.json")
	}
	if !activeFound {
		snapshot, snapshotErr := application.ports.Snapshot(ctx, location.Root, location.Head, configuration.MaxGitOutputBytes, configuration.MaxGitPaths)
		if snapshotErr != nil {
			return application.failure("L7-GIT-001", request.Command, "invalid", "cannot reconstruct idle Git status: "+bounded(snapshotErr.Error(), 512), "stabilize the Git worktree, then retry l7 status")
		}
		if !sameCandidate(location, snapshot.RepositoryLocation) {
			return application.failure("L7-GIT-002", request.Command, "changed", "Git identity changed during idle status reconstruction", "retry l7 status against a stable repository")
		}
		if !repositoryOutputFits(configuration.MaxCommandOutputBytes, snapshot, nil) {
			return application.result(domain.OutcomeBlocked, "L7-OUTPUT-001", string(request.Command), "bounded", "Git status exceeds the configured command-output limit", "narrow the active scope or explicitly raise the bounded output limit")
		}
		recheckedConfiguration, configFound, configErr := application.ports.LoadConfiguration(location.Root)
		_, recheckedActiveFound, activeErr := application.ports.LoadActive(location.CommonDir)
		if configErr != nil || !configFound || activeErr != nil || recheckedActiveFound || !sameConfiguration(configuration, recheckedConfiguration) {
			return application.failure("L7-STATE-004", request.Command, "changed", "configuration or active context changed during idle status reconstruction", "retry l7 status against stable local state")
		}
		outcome := domain.OutcomePass
		code := "L7-STATUS-000"
		state := "idle"
		message := "repository is adopted with no active change"
		next := "run l7 brief with an explicit risk tier and scope"
		if len(snapshot.ChangedPaths) != 0 {
			outcome = domain.OutcomeBlocked
			code = "L7-BRIEF-001"
			state = "idle-dirty"
			message = "repository has uncommitted changes but no active Level 7 change"
			next = "commit or restore current changes before running l7 brief"
		}
		result := application.result(outcome, code, string(request.Command), state, message, next)
		result.Repository = detailsForSnapshot(snapshot, "", 0, nil, nil)
		return result
	}
	if !application.executionPortsAvailable() {
		return application.result(domain.OutcomeBlocked, "L7-CAP-003", string(request.Command), "unavailable", "execution evidence is not configured in this build", "run l7 help")
	}
	change, blocked := application.loadExecutionContext(ctx, request.Command)
	if blocked != nil {
		return *blocked
	}
	if !sameStoredActive(change.active, initialActive) {
		return application.failure("L7-STATE-004", request.Command, "changed", "active context changed during status reconstruction", "retry l7 status against stable local state")
	}
	return application.statusFromContext(ctx, request, change)
}

func (application Application) lifecyclePortsAvailable() bool {
	return application.cwd != "" && application.ports.Locate != nil && application.ports.Snapshot != nil && application.ports.LoadConfiguration != nil && application.ports.AdoptConfiguration != nil && application.ports.LoadActive != nil && application.ports.SaveActive != nil && application.ports.Acquire != nil && application.ports.EnsureBrief != nil && application.ports.LoadBrief != nil
}

func (application Application) failure(code string, command domain.Command, state, message, next string) domain.Result {
	return application.result(domain.OutcomeFailed, code, string(command), state, message, next)
}

func repositoryDetails(location domain.RepositoryLocation) *domain.RepositoryDetails {
	return &domain.RepositoryDetails{Root: location.Root, CommonDir: location.CommonDir, Head: location.Head, Tree: location.Tree, DeclaredScope: []string{}, ChangedPaths: []string{}, ExpandedPaths: []string{}}
}

func detailsForSnapshot(snapshot domain.RepositorySnapshot, changeID string, tier domain.RiskTier, scope, expanded []string) *domain.RepositoryDetails {
	return &domain.RepositoryDetails{
		Root:          snapshot.Root,
		CommonDir:     snapshot.CommonDir,
		ChangeID:      changeID,
		Tier:          tier,
		Base:          snapshot.Base,
		Head:          snapshot.Head,
		Tree:          snapshot.Tree,
		DeclaredScope: append([]string{}, scope...),
		ChangedPaths:  append([]string{}, snapshot.ChangedPaths...),
		ExpandedPaths: append([]string{}, expanded...),
	}
}

func onlyPermitted(changed []string, permitted string) bool {
	if len(changed) == 0 {
		return true
	}
	return permitted != "" && len(changed) == 1 && changed[0] == permitted
}

func sameLocation(left, right domain.RepositoryLocation) bool {
	return left.Root == right.Root && left.CommonDir == right.CommonDir && left.Head == right.Head && left.Tree == right.Tree
}

func sameCandidate(location domain.RepositoryLocation, snapshot domain.RepositoryLocation) bool {
	return sameLocation(location, snapshot)
}

func sameSnapshot(left, right domain.RepositorySnapshot) bool {
	return sameLocation(left.RepositoryLocation, right.RepositoryLocation) && left.Base == right.Base && sameStrings(left.ChangedPaths, right.ChangedPaths)
}

func sameConfiguration(left, right domain.Configuration) bool {
	return left.LocalLifecycle == right.LocalLifecycle && left.MaxInputBytes == right.MaxInputBytes && left.MaxGitOutputBytes == right.MaxGitOutputBytes && left.MaxGitPaths == right.MaxGitPaths && left.MaxCommandOutputBytes == right.MaxCommandOutputBytes && left.MaxCommandSeconds == right.MaxCommandSeconds && left.Implementer == right.Implementer && left.Reviewer == right.Reviewer && sameStrings(left.ProtectedPaths, right.ProtectedPaths) && sameVerificationCommands(left.Verification, right.Verification)
}

func sameStoredActive(resolved, stored domain.ActiveChange) bool {
	if resolved.Kind == domain.ActiveBrief {
		return stored.Kind == domain.ActiveBrief && stored.ID == resolved.ID && stored.BriefPath == resolved.BriefPath && stored.Tier == 0 && stored.Base == "" && stored.Problem == "" && len(stored.Scope) == 0
	}
	return resolved.Kind == stored.Kind && resolved.ID == stored.ID && resolved.Tier == stored.Tier && resolved.Base == stored.Base && resolved.Problem == stored.Problem && resolved.BriefPath == stored.BriefPath && sameStrings(resolved.Scope, stored.Scope)
}

func sameBrief(left, right domain.ChangeBrief) bool {
	return left.ID == right.ID && left.Tier == right.Tier && left.Base == right.Base && left.Path == right.Path && left.Problem == right.Problem && sameStrings(left.Scope, right.Scope) && sameStrings(left.AcceptanceCriteria, right.AcceptanceCriteria) && sameStrings(left.Risks, right.Risks) && sameStrings(left.Rollback, right.Rollback)
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func touchesProtected(scope, additions []string) bool {
	protected := append(append([]string{}, builtinProtectedPaths...), additions...)
	for _, scoped := range scope {
		for _, control := range protected {
			if patternsOverlap(scoped, control) {
				return true
			}
		}
	}
	return false
}

func patternsOverlap(left, right string) bool {
	leftRecursive := strings.HasSuffix(left, "/**")
	rightRecursive := strings.HasSuffix(right, "/**")
	leftBase := strings.TrimSuffix(left, "/**")
	rightBase := strings.TrimSuffix(right, "/**")
	if !leftRecursive && !rightRecursive {
		return leftBase == rightBase
	}
	if leftRecursive && rightRecursive {
		return leftBase == rightBase || strings.HasPrefix(leftBase, rightBase+"/") || strings.HasPrefix(rightBase, leftBase+"/")
	}
	if leftRecursive {
		return rightBase == leftBase || strings.HasPrefix(rightBase, leftBase+"/")
	}
	return leftBase == rightBase || strings.HasPrefix(leftBase, rightBase+"/")
}

func requestBytes(request domain.Request) int {
	total := len(request.ChangeID) + len(request.Problem)
	for _, group := range [][]string{request.Scope, request.AcceptanceCriteria, request.Risks, request.Rollback} {
		for _, value := range group {
			total += len(value)
		}
	}
	return total
}

func repositoryOutputFits(limit int, snapshot domain.RepositorySnapshot, scope []string) bool {
	if limit < 1 {
		return false
	}
	remaining := limit - 4096
	for _, value := range []string{snapshot.Root, snapshot.CommonDir, snapshot.Base, snapshot.Head, snapshot.Tree} {
		if !consumeOutputBudget(&remaining, value) {
			return false
		}
	}
	for _, group := range [][]string{scope, snapshot.ChangedPaths} {
		for _, value := range group {
			if !consumeOutputBudget(&remaining, value) {
				return false
			}
		}
	}
	return remaining >= 0
}

func consumeOutputBudget(remaining *int, value string) bool {
	const expansionFactor = 12
	if *remaining < 64 || len(value) > (*remaining-64)/expansionFactor {
		return false
	}
	*remaining -= len(value)*expansionFactor + 64
	return true
}
