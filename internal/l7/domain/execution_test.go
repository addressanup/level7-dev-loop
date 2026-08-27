package domain

import "testing"

func TestProviderContractAndCrossProviderSelection(t *testing.T) {
	if !ProviderCodex.Valid() || !ProviderClaude.Valid() || Provider("shell").Valid() {
		t.Fatal("provider validation is not closed")
	}
	other, ok := OtherProvider(ProviderCodex)
	if !ok || other != ProviderClaude {
		t.Fatalf("OtherProvider(codex)=(%q,%v)", other, ok)
	}
	if _, ok := OtherProvider("unknown"); ok {
		t.Fatal("unknown provider has a fallback")
	}
}

func TestConventionalSubjectIsBoundedAndSingleLine(t *testing.T) {
	for _, value := range []string{"feat(cli): add execution", "fix: contain child process", "test(provider)!: reject drift"} {
		if !ConventionalSubject(value) {
			t.Fatalf("ConventionalSubject(%q)=false", value)
		}
	}
	for _, value := range []string{"feature without delimiter", "Feat: uppercase", "fix:  double space", "fix:\nattack", "x: short"} {
		if ConventionalSubject(value) {
			t.Fatalf("ConventionalSubject(%q)=true", value)
		}
	}
}

func TestTierThreeReviewerMustUseOtherProvider(t *testing.T) {
	codex := ProviderIdentity{Provider: ProviderCodex, Executable: "/bin/codex", Digest: "a"}
	secondCodex := ProviderIdentity{Provider: ProviderCodex, Executable: "/other/codex", Digest: "b"}
	claude := ProviderIdentity{Provider: ProviderClaude, Executable: "/bin/claude", Digest: "c"}
	if DistinctReviewer(TierHighRisk, codex, secondCodex) {
		t.Fatal("Tier 3 accepted the implementer provider as reviewer")
	}
	if !DistinctReviewer(TierHighRisk, codex, claude) {
		t.Fatal("Tier 3 rejected the other provider")
	}
	if !DistinctReviewer(TierProduct, codex, secondCodex) {
		t.Fatal("normal review rejected a distinct invocation identity")
	}
}
