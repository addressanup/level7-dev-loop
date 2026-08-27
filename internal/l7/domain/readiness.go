package domain

const ReadinessEvidenceSchema = 1

type ReadinessEvidence struct {
	ChangeID            string
	Tier                RiskTier
	Base                string
	Candidate           CandidateIdentity
	BriefCommit         string
	ConfigurationDigest string
	VerificationCommit  string
	ReviewCommit        string
	Scope               []string
	Checks              []CheckResult
	Owner               string
	Implementer         Provider
	Reviewer            Provider
	ReviewDecision      ReviewDecision
	BenchmarkRequired   bool
}

type ReadinessFacts struct {
	Evidence            ReadinessEvidence
	PlanCurrent         bool
	RepositoryClean     bool
	ApprovalCurrent     bool
	VerificationCurrent bool
	ReviewCurrent       bool
	AuditCurrent        bool
}

type ReadinessFinding struct {
	Code    string
	Message string
}

type ReadinessDecision struct {
	Ready    bool
	Findings []ReadinessFinding
}

func EvaluateReadiness(facts ReadinessFacts) ReadinessDecision {
	evidence := facts.Evidence
	findings := make([]ReadinessFinding, 0)
	add := func(code, message string) {
		findings = append(findings, ReadinessFinding{Code: code, Message: message})
	}
	if !safeReadinessToken(evidence.ChangeID, 64) || !evidence.Tier.Valid() || len(evidence.Scope) < 1 || len(evidence.Scope) > 100_000 {
		add("L7-READY-F001", "change identity, risk tier, or declared scope is invalid")
	}
	if !fullReadinessID(evidence.Base) || !fullReadinessID(evidence.Candidate.Commit) || !fullReadinessID(evidence.Candidate.Tree) || evidence.Candidate.Commit == evidence.Base {
		add("L7-READY-F002", "base or candidate Git identity is invalid")
	}
	if evidence.Tier != TierRoutine && !fullReadinessID(evidence.BriefCommit) {
		add("L7-READY-F003", "tracked brief identity is absent or invalid")
	}
	if !fullReadinessID(evidence.VerificationCommit) || !fullReadinessID(evidence.ReviewCommit) || !hexReadinessDigest(evidence.ConfigurationDigest) {
		add("L7-READY-F004", "verification, review, or configuration binding is invalid")
	}
	if !safeReadinessScope(evidence.Scope) {
		add("L7-READY-F005", "declared scope contains an unsafe or duplicate entry")
	}
	benchmarkPresent := false
	if !passingReadinessChecks(evidence.Checks, &benchmarkPresent) {
		add("L7-READY-F006", "verification checks are absent, failing, duplicate, or unsafe")
	}
	if !evidence.Implementer.Valid() || !evidence.Reviewer.Valid() || !evidence.ReviewDecision.Valid() || evidence.ReviewDecision != DecisionGO {
		add("L7-READY-F007", "implementer, reviewer, or review decision is invalid")
	}
	if !facts.PlanCurrent || !facts.RepositoryClean || !facts.VerificationCurrent {
		add("L7-READY-F008", "plan, clean-repository, or verification facts are stale")
	}
	if evidence.Tier == TierHighRisk {
		if !facts.ApprovalCurrent || !facts.AuditCurrent || !safeReadinessToken(evidence.Owner, 128) {
			add("L7-READY-F009", "Tier 3 owner approval or independent audit is not current")
		}
		if evidence.Implementer == evidence.Reviewer || evidence.Owner == string(evidence.Implementer) || evidence.Owner == string(evidence.Reviewer) {
			add("L7-READY-F010", "Tier 3 owner, implementer, and reviewer are not distinct")
		}
		if !evidence.BenchmarkRequired {
			add("L7-READY-F011", "Tier 3 readiness did not require benchmark evidence")
		}
	} else if !facts.ReviewCurrent {
		add("L7-READY-F012", "review is not current")
	}
	if evidence.BenchmarkRequired && !benchmarkPresent {
		add("L7-READY-F013", "required benchmark evidence is absent")
	}
	return ReadinessDecision{Ready: len(findings) == 0, Findings: findings}
}

func ReadinessEvidenceValid(evidence ReadinessEvidence) bool {
	facts := ReadinessFacts{
		Evidence: evidence, PlanCurrent: true, RepositoryClean: true,
		ApprovalCurrent:     evidence.Tier == TierHighRisk,
		VerificationCurrent: true, ReviewCurrent: evidence.Tier != TierHighRisk,
		AuditCurrent: evidence.Tier == TierHighRisk,
	}
	return EvaluateReadiness(facts).Ready
}

func fullReadinessID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func hexReadinessDigest(value string) bool {
	return len(value) == 64 && fullReadinessID(value)
}

func safeReadinessToken(value string, maximum int) bool {
	if len(value) < 1 || len(value) > maximum {
		return false
	}
	for _, character := range value {
		if character == 0x7f || character < 0x20 {
			return false
		}
	}
	return true
}

func safeReadinessScope(scope []string) bool {
	seen := make(map[string]bool, len(scope))
	for _, value := range scope {
		if !safeReadinessToken(value, 512) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func passingReadinessChecks(checks []CheckResult, benchmarkPresent *bool) bool {
	if len(checks) < 1 || len(checks) > 32 {
		return false
	}
	seen := make(map[string]bool, len(checks))
	for _, check := range checks {
		if !safeReadinessToken(check.Name, 64) || seen[check.Name] || !check.Passed || check.ExitCode != 0 {
			return false
		}
		seen[check.Name] = true
		*benchmarkPresent = *benchmarkPresent || check.Benchmark
	}
	return true
}
