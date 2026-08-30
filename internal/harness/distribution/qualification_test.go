package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestOfflineQualificationReportIsDeterministicAndReleaseBlocked(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	packages, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	lifecycle, err := qualifyLifecycleSet(packages)
	if err != nil {
		t.Fatal(err)
	}
	facts := offlineQualificationFacts{
		Descriptor: inputs.Descriptor, Packages: packages, RebuiltPackages: rebuilt, Lifecycle: lifecycle,
	}
	first, err := qualifyOfflinePackageSet(facts)
	if err != nil {
		t.Fatal(err)
	}
	second, err := qualifyOfflinePackageSet(facts)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, err := jsonBytes(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := jsonBytes(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("qualification report changed between evaluations:\n%s\n%s", firstJSON, secondJSON)
	}
	if first.Schema != 1 || first.Kind != offlineQualificationKind || first.Version != inputs.Descriptor.Version ||
		first.Channel != inputs.Descriptor.Channel || first.InternalStructure != qualificationPass ||
		first.Reproducibility != qualificationPass || first.FixtureLifecycle != qualificationPass ||
		first.Signing != qualificationNotRun || first.Publication != qualificationNotRun ||
		first.Authority != qualificationUnevaluated || first.ReleaseReady || len(first.Packages) != 2 {
		t.Fatalf("unexpected qualification report: %+v", first)
	}
	for index, host := range []string{"codex", "claude"} {
		qualified := first.Packages[index]
		if qualified.Host != host || qualified.Version != inputs.Descriptor.Version ||
			!sha256Pattern.MatchString(qualified.SourceDigest) ||
			qualified.ArchiveSHA256 != packages[index].ArchiveDigest ||
			qualified.CatalogPath != packages[index].CatalogPath || qualified.CatalogSHA256 != packages[index].CatalogDigest ||
			qualified.HostValidation != qualificationNotRun || qualified.MarketplaceLifecycle != qualificationNotRun ||
			qualified.PluginInstallation != qualificationNotRun || qualified.PluginActivation != qualificationNotRun ||
			qualified.BehavioralActivation != qualificationNotRun || qualified.ProviderExecution != qualificationNotRun ||
			qualified.SupportClaim != qualificationWithheld {
			t.Fatalf("unexpected %s qualification: %+v", host, qualified)
		}
	}
	if first.Packages[0].SourceDigest == first.Packages[1].SourceDigest ||
		first.Packages[0].ArchiveSHA256 == first.Packages[1].ArchiveSHA256 ||
		first.Packages[0].CatalogSHA256 == first.Packages[1].CatalogSHA256 {
		t.Fatalf("host identities are not independent: %+v", first.Packages)
	}
	encoded := string(firstJSON)
	for _, required := range []string{
		`"internal_structure": "PASS"`, `"reproducibility": "PASS"`, `"fixture_lifecycle": "PASS"`,
		`"host_package_validation": "NOT_RUN"`, `"marketplace_lifecycle": "NOT_RUN"`,
		`"plugin_installation": "NOT_RUN"`, `"plugin_activation": "NOT_RUN"`,
		`"behavioral_activation": "NOT_RUN"`, `"provider_execution": "NOT_RUN"`,
		`"support_claim": "WITHHELD"`, `"signing": "NOT_RUN"`, `"publication": "NOT_RUN"`,
		`"authority": "NOT_EVALUATED"`, `"release_ready": false`,
	} {
		if !strings.Contains(encoded, required) {
			t.Fatalf("qualification report lacks %q: %s", required, encoded)
		}
	}
	if strings.Contains(encoded, `"release_ready": true`) || strings.Contains(encoded, `"authority": "APPROVED"`) {
		t.Fatalf("qualification report made a release or authority claim: %s", encoded)
	}
}

func TestOfflineQualificationFailsClosed(t *testing.T) {
	inputs, err := loadInputs(distributionRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	packages, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	rebuilt, err := buildPackages(inputs)
	if err != nil {
		t.Fatal(err)
	}
	lifecycle, err := qualifyLifecycleSet(packages)
	if err != nil {
		t.Fatal(err)
	}
	base := offlineQualificationFacts{
		Descriptor: inputs.Descriptor, Packages: packages, RebuiltPackages: rebuilt, Lifecycle: lifecycle,
	}
	tests := []struct {
		name   string
		mutate func(*offlineQualificationFacts)
	}{
		{name: "lifecycle not qualified", mutate: func(value *offlineQualificationFacts) { value.Lifecycle = lifecycleQualification{} }},
		{name: "lifecycle identity substitution", mutate: func(value *offlineQualificationFacts) {
			value.Lifecycle.Packages[1].CatalogSHA256 = value.Lifecycle.Packages[0].CatalogSHA256
		}},
		{name: "missing host", mutate: func(value *offlineQualificationFacts) { value.Packages = value.Packages[:1] }},
		{name: "reordered hosts", mutate: func(value *offlineQualificationFacts) {
			value.Packages[0], value.Packages[1] = value.Packages[1], value.Packages[0]
		}},
		{name: "descriptor development channel", mutate: func(value *offlineQualificationFacts) { value.Descriptor.Channel = "development" }},
		{name: "descriptor prerelease version", mutate: func(value *offlineQualificationFacts) { value.Descriptor.Version = "0.1.0-dev.6" }},
		{name: "descriptor stable version substitution", mutate: func(value *offlineQualificationFacts) { value.Descriptor.Version = "0.1.2" }},
		{name: "archive substitution", mutate: func(value *offlineQualificationFacts) {
			value.Packages[0].ArchiveDigest = strings.Repeat("0", 64)
		}},
		{name: "catalog substitution", mutate: func(value *offlineQualificationFacts) {
			value.Packages[0].CatalogDigest = strings.Repeat("0", 64)
		}},
		{name: "source relabeling", mutate: func(value *offlineQualificationFacts) {
			value.Packages[0].SourceDigest = strings.Repeat("0", 64)
		}},
		{name: "non-reproducible catalog", mutate: func(value *offlineQualificationFacts) {
			value.RebuiltPackages[0].Catalog = append(value.RebuiltPackages[0].Catalog, '\n')
			value.RebuiltPackages[0].CatalogDigest = sha256Hex(value.RebuiltPackages[0].Catalog)
		}},
		{name: "cross-host identity", mutate: func(value *offlineQualificationFacts) {
			value.Packages[1].Catalog = append([]byte{}, value.Packages[0].Catalog...)
			value.Packages[1].CatalogDigest = value.Packages[0].CatalogDigest
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			facts := cloneOfflineQualificationFacts(base)
			test.mutate(&facts)
			if report, err := qualifyOfflinePackageSet(facts); err == nil {
				t.Fatalf("unsafe facts passed with report: %+v", report)
			}
		})
	}
}

func cloneOfflineQualificationFacts(original offlineQualificationFacts) offlineQualificationFacts {
	clone := original
	clone.Descriptor.Keywords = append([]string{}, original.Descriptor.Keywords...)
	clone.Descriptor.Skills = append([]string{}, original.Descriptor.Skills...)
	clone.Packages = cloneOfflineQualificationPackages(original.Packages)
	clone.RebuiltPackages = cloneOfflineQualificationPackages(original.RebuiltPackages)
	clone.Lifecycle.Packages = append([]lifecycleQualificationPackage{}, original.Lifecycle.Packages...)
	return clone
}

func cloneOfflineQualificationPackages(original []builtPackage) []builtPackage {
	cloned := make([]builtPackage, len(original))
	for index, built := range original {
		cloned[index] = built
		cloned[index].Entries = cloneEntries(built.Entries)
		cloned[index].Archive = append([]byte{}, built.Archive...)
		cloned[index].Catalog = append([]byte{}, built.Catalog...)
	}
	return cloned
}
