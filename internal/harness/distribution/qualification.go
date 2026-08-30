package main

import (
	"errors"
	"fmt"
)

const (
	offlineQualificationKind = "offline-package-qualification"
	qualificationPass        = "PASS"
	qualificationSmokeTested = "SMOKE_TESTED"
	qualificationNotRun      = "NOT_RUN"
	qualificationWithheld    = "WITHHELD"
	qualificationUnevaluated = "NOT_EVALUATED"
)

type offlineQualificationFacts struct {
	Descriptor      packageDescriptor
	Packages        []builtPackage
	RebuiltPackages []builtPackage
	Lifecycle       lifecycleQualification
}

type offlineQualificationPackage struct {
	Host                 string `json:"host"`
	Version              string `json:"version"`
	SourceDigest         string `json:"source_digest"`
	ArchiveSHA256        string `json:"archive_sha256"`
	CatalogPath          string `json:"catalog_path"`
	CatalogSHA256        string `json:"catalog_sha256"`
	HostValidation       string `json:"host_package_validation"`
	MarketplaceLifecycle string `json:"marketplace_lifecycle"`
	PluginInstallation   string `json:"plugin_installation"`
	PluginActivation     string `json:"plugin_activation"`
	BehavioralActivation string `json:"behavioral_activation"`
	ProviderExecution    string `json:"provider_execution"`
	SupportClaim         string `json:"support_claim"`
}

type offlineQualificationReport struct {
	Schema            int                           `json:"schema"`
	Kind              string                        `json:"kind"`
	Version           string                        `json:"version"`
	Channel           string                        `json:"channel"`
	Packages          []offlineQualificationPackage `json:"packages"`
	InternalStructure string                        `json:"internal_structure"`
	Reproducibility   string                        `json:"reproducibility"`
	FixtureLifecycle  string                        `json:"fixture_lifecycle"`
	Signing           string                        `json:"signing"`
	Publication       string                        `json:"publication"`
	Authority         string                        `json:"authority"`
	ReleaseReady      bool                          `json:"release_ready"`
}

// qualifyOfflinePackageSet reports only the facts established by the inert
// package builder and disposable lifecycle fixtures. It deliberately accepts
// no release authority or external-gate input and can never authorize release.
func qualifyOfflinePackageSet(facts offlineQualificationFacts) (offlineQualificationReport, error) {
	if err := validateDescriptor(facts.Descriptor); err != nil {
		return offlineQualificationReport{}, fmt.Errorf("qualification descriptor: %w", err)
	}
	if err := validateQualificationPackageOrder(facts.Packages); err != nil {
		return offlineQualificationReport{}, err
	}
	if err := validateQualificationPackageOrder(facts.RebuiltPackages); err != nil {
		return offlineQualificationReport{}, fmt.Errorf("rebuilt package set: %w", err)
	}
	if err := compareBuilds(facts.Packages, facts.RebuiltPackages); err != nil {
		return offlineQualificationReport{}, fmt.Errorf("offline package reproducibility: %w", err)
	}
	if err := validateLifecycleQualification(facts.Lifecycle, facts.Packages); err != nil {
		return offlineQualificationReport{}, fmt.Errorf("offline fixture lifecycle: %w", err)
	}

	qualified := make([]offlineQualificationPackage, 0, len(facts.Packages))
	seenSource := make(map[string]bool, len(facts.Packages))
	seenArchive := make(map[string]bool, len(facts.Packages))
	seenCatalog := make(map[string]bool, len(facts.Packages))
	for _, built := range facts.Packages {
		if err := validateBuiltPackage(built); err != nil {
			return offlineQualificationReport{}, fmt.Errorf("%s package structure: %w", built.Host, err)
		}
		if built.Version != facts.Descriptor.Version {
			return offlineQualificationReport{}, fmt.Errorf("%s package version does not match the descriptor", built.Host)
		}
		if !sha256Pattern.MatchString(built.CatalogDigest) || sha256Hex(built.Catalog) != built.CatalogDigest {
			return offlineQualificationReport{}, fmt.Errorf("%s catalog digest mismatch", built.Host)
		}

		metadataData, count := packageEntry(built.Entries, "DISTRIBUTION.json")
		if count != 1 {
			return offlineQualificationReport{}, fmt.Errorf("%s package lacks one distribution identity", built.Host)
		}
		var metadata distributionMetadata
		if err := decodeStrict(metadataData, &metadata); err != nil {
			return offlineQualificationReport{}, fmt.Errorf("decode %s distribution identity: %w", built.Host, err)
		}
		if metadata.Schema != 2 || metadata.Host != built.Host || metadata.Version != built.Version ||
			metadata.Channel != facts.Descriptor.Channel || metadata.CatalogPath != built.CatalogPath ||
			metadata.CatalogSHA256 != built.CatalogDigest || metadata.SourceDigest != built.SourceDigest ||
			!sha256Pattern.MatchString(metadata.SourceDigest) ||
			metadata.ActualHostGate != qualificationSmokeTested || metadata.SupportClaim != qualificationWithheld {
			return offlineQualificationReport{}, fmt.Errorf("%s distribution identity exceeds the offline qualification boundary", built.Host)
		}

		compatibilityData, count := packageEntry(built.Entries, "COMPATIBILITY.json")
		if count != 1 {
			return offlineQualificationReport{}, fmt.Errorf("%s package lacks one compatibility identity", built.Host)
		}
		var compatibility compatibilityProjection
		if err := decodeStrict(compatibilityData, &compatibility); err != nil {
			return offlineQualificationReport{}, fmt.Errorf("decode %s compatibility identity: %w", built.Host, err)
		}
		if compatibility.PackageVersion != built.Version || compatibility.Entry.Host != built.Host ||
			compatibility.Entry.ProviderExecution != qualificationNotRun ||
			compatibility.Entry.ActualHostLifecycle != qualificationSmokeTested ||
			compatibility.Entry.SupportClaim != qualificationWithheld {
			return offlineQualificationReport{}, fmt.Errorf("%s compatibility identity exceeds the offline qualification boundary", built.Host)
		}
		if seenSource[metadata.SourceDigest] || seenArchive[built.ArchiveDigest] || seenCatalog[built.CatalogDigest] {
			return offlineQualificationReport{}, errors.New("host package identities are not independent")
		}
		seenSource[metadata.SourceDigest] = true
		seenArchive[built.ArchiveDigest] = true
		seenCatalog[built.CatalogDigest] = true

		qualified = append(qualified, offlineQualificationPackage{
			Host: built.Host, Version: built.Version, SourceDigest: metadata.SourceDigest,
			ArchiveSHA256: built.ArchiveDigest, CatalogPath: built.CatalogPath, CatalogSHA256: built.CatalogDigest,
			HostValidation: qualificationNotRun, MarketplaceLifecycle: qualificationNotRun,
			PluginInstallation: qualificationNotRun, PluginActivation: qualificationNotRun,
			BehavioralActivation: qualificationNotRun, ProviderExecution: qualificationNotRun,
			SupportClaim: qualificationWithheld,
		})
	}

	return offlineQualificationReport{
		Schema: 1, Kind: offlineQualificationKind, Version: facts.Descriptor.Version,
		Channel: facts.Descriptor.Channel, Packages: qualified,
		InternalStructure: qualificationPass, Reproducibility: qualificationPass, FixtureLifecycle: qualificationPass,
		Signing: qualificationNotRun, Publication: qualificationNotRun, Authority: qualificationUnevaluated,
		ReleaseReady: false,
	}, nil
}

func validateQualificationPackageOrder(packages []builtPackage) error {
	if len(packages) != 2 {
		return errors.New("offline qualification requires exactly two host packages")
	}
	for index, host := range []string{"codex", "claude"} {
		if packages[index].Host != host {
			return fmt.Errorf("offline qualification package %d must be %s", index, host)
		}
		if err := validateBuiltPackage(packages[index]); err != nil {
			return fmt.Errorf("%s package structure: %w", host, err)
		}
	}
	if packages[0].Version != packages[1].Version {
		return errors.New("offline qualification package versions diverge")
	}
	return nil
}
