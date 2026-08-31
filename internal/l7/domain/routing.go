package domain

import "strings"

// Route selects the least-cost qualified model deterministically. Callers add
// the observation timestamp after the pure decision is returned.
func Route(task TaskProfile, snapshots []ProviderSnapshot) RouteDecision {
	decision := RouteDecision{Schema: OrchestrationSchema, TaskID: task.ID, Policy: "balanced", Candidates: []RouteCandidate{}, Fallbacks: []string{}, Escalations: []string{}, Next: "resolve provider capability blockers"}
	for _, snapshot := range snapshots {
		for _, model := range snapshot.Models {
			candidate := qualifyRouteCandidate(task, snapshot, model)
			decision.Candidates = append(decision.Candidates, candidate)
		}
	}
	sortRouteCandidates(decision.Candidates)
	for _, candidate := range decision.Candidates {
		if !candidate.Qualified {
			continue
		}
		if decision.ProviderID == "" {
			decision.ProviderID = candidate.ProviderID
			decision.ModelID = candidate.ModelID
			decision.Effort = candidate.Effort
			decision.Next = "start the selected provider assignment"
			continue
		}
		decision.Fallbacks = append(decision.Fallbacks, candidate.ProviderID+"/"+candidate.ModelID+"@"+string(candidate.Effort))
	}
	return decision
}

func qualifyRouteCandidate(task TaskProfile, snapshot ProviderSnapshot, model ModelCapability) RouteCandidate {
	candidate := RouteCandidate{ProviderID: snapshot.ID, ModelID: model.ID, Qualified: true, Score: 1_000_000, Reasons: []string{}}
	if snapshot.Authentication != AuthAuthenticated {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "provider is not authenticated")
	}
	if !model.Verified {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "model capability is not verified")
	}
	if task.ContextTokens > 0 && model.ContextWindow < task.ContextTokens {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "context window is too small")
	}
	if task.NeedsTools && !model.SupportsTools {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "tool use is unavailable")
	}
	if task.NeedsEditing && !model.SupportsEditing {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "workspace editing is unavailable")
	}
	if task.NeedsResume && !model.SupportsResume {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "session resume is unavailable")
	}
	for _, language := range task.Languages {
		if !supportsLanguage(model.Languages, language) {
			candidate.Qualified = false
			candidate.Reasons = append(candidate.Reasons, "language "+language+" is not advertised")
		}
	}
	if task.IndependentReview && task.ImplementerProvider == snapshot.ID && task.ImplementerModel == model.ID {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "implementer model cannot audit itself")
	}
	effort, ok := effortForTask(task, model.Efforts)
	if !ok {
		candidate.Qualified = false
		candidate.Reasons = append(candidate.Reasons, "model does not advertise the required effort")
	} else {
		candidate.Effort = effort
	}
	if candidate.Qualified {
		candidate.Score = model.CostClass*100 + model.LatencyClass*20 + effortRank(effort)
		candidate.Reasons = append(candidate.Reasons, "qualified by auth, context, tools, risk, and effort")
	}
	return candidate
}

func effortForTask(task TaskProfile, advertised []ReasoningEffort) (ReasoningEffort, bool) {
	minimum := 2
	switch task.Complexity {
	case ComplexityC1:
		minimum = 2
	case ComplexityC2:
		minimum = 3
	case ComplexityC3, ComplexityC4:
		minimum = 4
	default:
		return "", false
	}
	if (task.RiskTier == TierHighRisk || task.PriorFailures > 0 || highEffortWork(task.WorkKinds)) && minimum < 4 {
		minimum = 4
	}
	selected := ReasoningEffort("")
	selectedRank := 100
	highest := ReasoningEffort("")
	highestRank := -1
	for _, effort := range advertised {
		rank := effortRank(effort)
		if rank < 0 {
			continue
		}
		if rank > highestRank {
			highest, highestRank = effort, rank
		}
		if rank >= minimum && rank < selectedRank {
			selected, selectedRank = effort, rank
		}
	}
	if task.Complexity == ComplexityC4 && highestRank >= minimum {
		return highest, true
	}
	return selected, selected != ""
}

func supportsLanguage(advertised []string, requested string) bool {
	requested = strings.ToLower(strings.TrimSpace(requested))
	if requested == "" {
		return false
	}
	for _, language := range advertised {
		language = strings.ToLower(strings.TrimSpace(language))
		if language == "*" || language == requested {
			return true
		}
	}
	return false
}

func highEffortWork(kinds []string) bool {
	for _, kind := range kinds {
		switch strings.ToLower(strings.TrimSpace(kind)) {
		case "security", "architecture", "release":
			return true
		}
	}
	return false
}

func effortRank(effort ReasoningEffort) int {
	switch effort {
	case EffortNone:
		return 0
	case EffortMinimal:
		return 1
	case EffortLow:
		return 2
	case EffortMedium:
		return 3
	case EffortHigh:
		return 4
	case EffortXHigh:
		return 5
	case EffortMax:
		return 6
	case EffortUltra:
		return 7
	default:
		return -1
	}
}

func sortRouteCandidates(candidates []RouteCandidate) {
	for index := 1; index < len(candidates); index++ {
		value := candidates[index]
		cursor := index - 1
		for cursor >= 0 && routeCandidateAfter(candidates[cursor], value) {
			candidates[cursor+1] = candidates[cursor]
			cursor--
		}
		candidates[cursor+1] = value
	}
}

func routeCandidateAfter(left, right RouteCandidate) bool {
	if left.Qualified != right.Qualified {
		return !left.Qualified
	}
	if left.Score != right.Score {
		return left.Score > right.Score
	}
	if left.ProviderID != right.ProviderID {
		return left.ProviderID > right.ProviderID
	}
	return left.ModelID > right.ModelID
}
