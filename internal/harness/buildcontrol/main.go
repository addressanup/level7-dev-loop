package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const buildControlVersion = "wave-02-v1"

type successSource struct {
	id   string
	path string
}

var successSources = []successSource{
	{"requirements", "docs/artifacts/requirements.md"},
	{"backlog", "docs/artifacts/feature-backlog.md"},
	{"support", "harness/support-matrix.tsv"},
	{"dispositions", "harness/prototype-dispositions.tsv"},
	{"phase", "harness/phases.tsv"},
	{"ownership", "harness/control-ownership.tsv"},
	{"orchestration", "docs/artifacts/orchestration-plan.md"},
	{"modules", "harness/modules.lock.tsv"},
	{"imports", "harness/import-boundaries.tsv"},
}

func main() {
	if len(os.Args) != 1 {
		printFindings([]finding{newFinding("BCTL-001", "arguments", "build-control accepts no arguments", "invoke the fixed command without options")})
		os.Exit(1)
	}
	root, findings := resolveRoot()
	if len(findings) != 0 {
		printFindings(findings)
		os.Exit(1)
	}
	output, findings := runController(root)
	if len(findings) != 0 {
		printFindings(findings)
		os.Exit(1)
	}
	fmt.Println(output)
}

func runController(root string) (string, []finding) {
	return runControllerWithSkillInventoryHook(root, nil)
}

func runControllerWithSkillInventoryHook(root string, beforeInventoryRead func()) (string, []finding) {
	var findings []finding
	trace, traceFindings := checkTrace(root)
	prototypeCount, claimFindings := checkClaimsWithSkillInventoryHook(root, beforeInventoryRead)
	policy, policyFindings := checkPolicy(root)
	ownershipCount, ownershipFindings := checkOwnership(root)
	findings = appendFindings(findings, traceFindings...)
	findings = appendFindings(findings, claimFindings...)
	findings = appendFindings(findings, policyFindings...)
	findings = appendFindings(findings, ownershipFindings...)
	sourceDigests, sourceFindings := loadSuccessSourceDigests(root)
	findings = appendFindings(findings, sourceFindings...)
	if len(findings) != 0 {
		return "", findings
	}
	return formatSuccess(trace, prototypeCount, policy, ownershipCount, sourceDigests), nil
}

func formatSuccess(trace traceResult, prototypeCount int, policy policyResult, ownershipCount int, sourceDigests string) string {
	return fmt.Sprintf("PASS rule=BCTL-000 gate_version=%s source_sha256=%s phase=%s checkpoint=%s requirements=%d allocation=v1.0:%d,v1.x:%d,later:%d prototypes=%d ownership=%d files=%d changed=%d",
		buildControlVersion,
		sourceDigests,
		policy.phase,
		policy.checkpoint,
		trace.total,
		trace.allocations["V1.0"],
		trace.allocations["V1.x"],
		trace.allocations["Later"],
		prototypeCount,
		ownershipCount,
		policy.files,
		policy.changed,
	)
}

func loadSuccessSourceDigests(root string) (string, []finding) {
	phase, findings := loadValidatedActivePhase(root)
	sources := append([]successSource(nil), successSources...)
	if len(findings) == 0 {
		sources = append(sources,
			successSource{"paths", phase.pathPolicy},
			successSource{"base", phase.baseManifest},
		)
		if phase.phase == "wave-02" {
			sources = append(sources,
				successSource{"approval", "docs/artifacts/wave-02-approval.md"},
				successSource{"contract", "docs/artifacts/wave-02-change-contract.md"},
				successSource{"specification", "docs/artifacts/wave-02-specification.md"},
				successSource{"design", "docs/artifacts/wave-02-design.md"},
				successSource{"predecessor-evidence", "docs/artifacts/wave-01-evidence.md"},
				successSource{"predecessor-audit", "docs/artifacts/wave-01-audit.md"},
			)
			if wave02CandidatePresent(root) {
				sources = append(sources, successSource{"candidate", wave02CandidateManifest})
			}
		}
	}
	parts := make([]string, 0, len(sources))
	for _, source := range sources {
		data, readFindings := readStrictFile(root, source.path)
		findings = appendFindings(findings, readFindings...)
		if len(readFindings) == 0 {
			parts = append(parts, source.id+":"+fileSHA256(data))
		}
	}
	return strings.Join(parts, ","), findings
}

func resolveRoot() (string, []finding) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", []finding{newFinding("BCTL-002", "cwd", err.Error(), "run from the canonical repository root")}
	}
	absolute, err := filepath.Abs(workingDirectory)
	if err != nil {
		return "", []finding{newFinding("BCTL-002", "cwd", err.Error(), "run from the canonical repository root")}
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", []finding{newFinding("BCTL-003", absolute, err.Error(), "use a canonical root with no symlink component")}
	}
	if resolved != absolute {
		return "", []finding{newFinding("BCTL-003", absolute, "repository root contains a symlink component", "use the canonical repository path")}
	}
	for _, marker := range []string{"go.mod", "AGENTS.md", "harness/phases.tsv"} {
		info, statErr := os.Lstat(filepath.Join(resolved, marker))
		if statErr != nil || !info.Mode().IsRegular() {
			return "", []finding{newFinding("BCTL-004", marker, "repository root marker is missing or not regular", "run from the canonical repository root")}
		}
	}
	return resolved, nil
}
