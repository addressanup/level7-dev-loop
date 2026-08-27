package domain

type RiskTier uint8

const (
	TierRoutine  RiskTier = 1
	TierProduct  RiskTier = 2
	TierHighRisk RiskTier = 3
)

func (tier RiskTier) Valid() bool {
	return tier >= TierRoutine && tier <= TierHighRisk
}

type LifecycleState string

const (
	StatePlanned                  LifecycleState = "planned"
	StateAwaitingOwnerApproval    LifecycleState = "awaiting-owner-approval"
	StateBuilding                 LifecycleState = "building"
	StateVerified                 LifecycleState = "verified"
	StateAwaitingIndependentAudit LifecycleState = "awaiting-independent-audit"
	StateReviewed                 LifecycleState = "reviewed"
	StateReady                    LifecycleState = "ready"
)

type Transition struct {
	State  LifecycleState
	Action string
}

func NextTransition(tier RiskTier, state LifecycleState) (Transition, bool) {
	if !tier.Valid() {
		return Transition{}, false
	}
	if tier == TierHighRisk {
		switch state {
		case StatePlanned:
			return Transition{State: StateAwaitingOwnerApproval, Action: "request explicit accountable-owner approval"}, true
		case StateAwaitingOwnerApproval:
			return Transition{State: StateBuilding, Action: "record current owner approval, then implement the approved scope"}, true
		case StateBuilding:
			return Transition{State: StateVerified, Action: "finish the scoped implementation and verify the exact Git candidate"}, true
		case StateVerified:
			return Transition{State: StateAwaitingIndependentAudit, Action: "request an independent read-only audit of the exact Git candidate"}, true
		case StateAwaitingIndependentAudit:
			return Transition{State: StateReviewed, Action: "record the current independent audit decision"}, true
		case StateReviewed:
			return Transition{State: StateReady, Action: "evaluate exact-candidate merge readiness"}, true
		case StateReady:
			return Transition{State: StateReady, Action: "request confirmation before merging the exact Git candidate"}, true
		default:
			return Transition{}, false
		}
	}

	switch state {
	case StatePlanned:
		return Transition{State: StateBuilding, Action: "implement the declared scope"}, true
	case StateBuilding:
		return Transition{State: StateVerified, Action: "finish the scoped implementation and run relevant verification"}, true
	case StateVerified:
		return Transition{State: StateReviewed, Action: "review the exact Git candidate"}, true
	case StateReviewed:
		return Transition{State: StateReady, Action: "evaluate exact-candidate merge readiness"}, true
	case StateReady:
		return Transition{State: StateReady, Action: "request confirmation before merging the exact Git candidate"}, true
	default:
		return Transition{}, false
	}
}

func TransitionAllowed(tier RiskTier, from, to LifecycleState) bool {
	next, ok := NextTransition(tier, from)
	if ok && next.State == to {
		return true
	}
	return to == StateBuilding && (from == StateVerified || from == StateReviewed || from == StateAwaitingIndependentAudit || from == StateReady)
}

type LifecycleFacts struct {
	Tier                    RiskTier
	PlanPresent             bool
	OwnerApprovalCurrent    bool
	WorkStarted             bool
	VerificationCurrent     bool
	ReviewCurrent           bool
	IndependentAuditCurrent bool
	ReadyCurrent            bool
	AssuranceRejected       bool
	AssuranceStale          bool
}

type RepositoryLocation struct {
	Root      string
	CommonDir string
	Head      string
	Tree      string
}

type RepositorySnapshot struct {
	RepositoryLocation
	Base         string
	ChangedPaths []string
}

type Configuration struct {
	LocalLifecycle        bool
	Verification          []VerificationCommand
	MaxInputBytes         int
	MaxGitOutputBytes     int
	MaxGitPaths           int
	MaxCommandOutputBytes int
	MaxCommandSeconds     int
	ProtectedPaths        []string
	Implementer           Provider
	Reviewer              Provider
}

type ChangeBrief struct {
	ID                 string
	Tier               RiskTier
	Base               string
	Path               string
	Problem            string
	Scope              []string
	AcceptanceCriteria []string
	Risks              []string
	Rollback           []string
}

type ActiveKind string

const (
	ActiveTierOne ActiveKind = "tier-1"
	ActiveBrief   ActiveKind = "brief"
)

type ActiveChange struct {
	Kind      ActiveKind
	ID        string
	Tier      RiskTier
	Base      string
	Problem   string
	Scope     []string
	BriefPath string
}

func ScopeContains(scope []string, relative string) bool {
	for _, declared := range scope {
		if declared == relative {
			return true
		}
		if hasSuffix(declared, "/**") {
			prefix := declared[:len(declared)-2]
			if hasPrefix(relative, prefix) && len(relative) > len(prefix) {
				return true
			}
		}
	}
	return false
}

func ExpandedPaths(scope, changed []string, permitted []string) []string {
	expanded := make([]string, 0)
	for _, relative := range changed {
		if !ScopeContains(scope, relative) && !ScopeContains(permitted, relative) {
			expanded = append(expanded, relative)
		}
	}
	return expanded
}

func hasPrefix(value, prefix string) bool {
	return len(value) >= len(prefix) && value[:len(prefix)] == prefix
}

func hasSuffix(value, suffix string) bool {
	return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
}

func DeriveLifecycle(facts LifecycleFacts) (LifecycleState, bool) {
	if !facts.Tier.Valid() || !facts.PlanPresent || lifecycleFactsConflict(facts) {
		return "", false
	}
	if facts.AssuranceRejected || facts.AssuranceStale {
		return StateBuilding, true
	}
	if facts.Tier == TierHighRisk && !facts.OwnerApprovalCurrent {
		return StateAwaitingOwnerApproval, true
	}
	if !facts.WorkStarted {
		if facts.Tier == TierHighRisk {
			return StateBuilding, true
		}
		return StatePlanned, true
	}
	if !facts.VerificationCurrent {
		return StateBuilding, true
	}
	if facts.Tier == TierHighRisk && !facts.IndependentAuditCurrent {
		return StateAwaitingIndependentAudit, true
	}
	if facts.Tier != TierHighRisk && !facts.ReviewCurrent {
		return StateVerified, true
	}
	if !facts.ReadyCurrent {
		return StateReviewed, true
	}
	return StateReady, true
}

func lifecycleFactsConflict(facts LifecycleFacts) bool {
	if facts.ReadyCurrent && (!facts.VerificationCurrent || (facts.Tier == TierHighRisk && !facts.IndependentAuditCurrent) || (facts.Tier != TierHighRisk && !facts.ReviewCurrent)) {
		return true
	}
	if facts.ReviewCurrent && !facts.VerificationCurrent {
		return true
	}
	if facts.IndependentAuditCurrent && (facts.Tier != TierHighRisk || !facts.VerificationCurrent) {
		return true
	}
	if (facts.VerificationCurrent || facts.ReviewCurrent || facts.IndependentAuditCurrent || facts.ReadyCurrent) && !facts.WorkStarted {
		return true
	}
	if facts.Tier != TierHighRisk && facts.OwnerApprovalCurrent {
		return true
	}
	if facts.Tier == TierHighRisk && !facts.OwnerApprovalCurrent && (facts.WorkStarted || facts.VerificationCurrent || facts.ReviewCurrent || facts.IndependentAuditCurrent || facts.ReadyCurrent) {
		return true
	}
	return facts.AssuranceRejected && facts.AssuranceStale
}
