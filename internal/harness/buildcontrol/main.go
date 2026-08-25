package main

import (
	"fmt"
	"os"
	"path/filepath"
)

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
	trace, traceFindings := checkTrace(root)
	prototypeCount, claimFindings := checkClaims(root)
	findings = append(findings, traceFindings...)
	findings = append(findings, claimFindings...)
	if len(findings) != 0 {
		printFindings(findings)
		os.Exit(1)
	}
	fmt.Printf("PASS rule=BCTL-000 phase=wave-01 requirements=%d allocation=v1.0:%d,v1.x:%d,later:%d prototypes=%d\n",
		trace.total,
		trace.allocations["V1.0"],
		trace.allocations["V1.x"],
		trace.allocations["Later"],
		prototypeCount,
	)
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
