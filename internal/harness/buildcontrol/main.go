package main

import (
	"flag"
	"fmt"
	"os"
)

const buildControlVersion = "solo-fast-v2"

func main() {
	repository := flag.String("repo", ".", "repository to evaluate")
	base := flag.String("base", os.Getenv("L7_BASE_REF"), "base Git revision")
	head := flag.String("head", envOr("L7_HEAD_REF", "HEAD"), "candidate Git revision")
	changeID := flag.String("change", os.Getenv("L7_CHANGE_ID"), "change ID")
	tier := flag.Int("tier", envInt("L7_RISK_TIER"), "risk tier for artifact-free Tier 1 work")
	assurance := flag.String("assurance", envOr("L7_ASSURANCE_MODE", string(assuranceSolo)), "assurance mode: solo or team")
	scope := flag.String("scope", os.Getenv("L7_SCOPE"), "comma-separated Tier 1 scope")
	requireReady := flag.Bool("require-ready", envBool("L7_REQUIRE_READY"), "require merge-ready state")
	flag.Parse()
	if flag.NArg() != 0 {
		printFindings([]finding{newFinding("BCTL-001", "arguments", "unexpected positional arguments", "use documented build-control flags")})
		os.Exit(1)
	}

	report, findings := runController(controllerOptions{
		Root:            *repository,
		BaseRef:         *base,
		HeadRef:         *head,
		ChangeID:        *changeID,
		Tier:            riskTier(*tier),
		Assurance:       assuranceMode(*assurance),
		TierOneScope:    splitScope(*scope),
		RequireReady:    *requireReady,
		VerifiedRef:     os.Getenv("L7_VERIFIED_REF"),
		ReviewRef:       os.Getenv("L7_REVIEW_REF"),
		AuditRequestRef: os.Getenv("L7_AUDIT_REQUEST_REF"),
		ReadyRef:        os.Getenv("L7_READY_REF"),
	})
	if len(findings) != 0 {
		printFindings(findings)
		os.Exit(1)
	}
	fmt.Println(report.String())
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
