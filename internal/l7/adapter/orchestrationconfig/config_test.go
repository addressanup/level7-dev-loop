package orchestrationconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/adapter/localfile"
	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestDefaultIsValidAndEffectsAreOff(t *testing.T) {
	configuration := Default()
	if err := configuration.Validate(); err != nil {
		t.Fatal(err)
	}
	if configuration.Features.Orchestration || configuration.Features.Sync || configuration.Features.CyberActive || configuration.Features.Headless {
		t.Fatal("default orchestration configuration enabled an effect")
	}
	if len(configuration.Providers) != 2 || configuration.Providers[0].Kind != domain.ProviderKindCodexAppServer || configuration.Providers[1].Kind != domain.ProviderKindClaudeCLI {
		t.Fatalf("unexpected native providers: %#v", configuration.Providers)
	}
}

func TestCreateAppliedIsExplicitAndDoesNotOverwrite(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	configuration, changed, err := Create(root, true)
	if err != nil || !changed || !configuration.Features.Orchestration || !configuration.Features.Sync || configuration.Features.CyberActive || configuration.Features.Headless {
		t.Fatalf("configuration=%#v changed=%t err=%v", configuration, changed, err)
	}
	data, err := os.ReadFile(Path(root))
	if err != nil || strings.Contains(string(data), "api_key") || strings.Contains(string(data), "Bearer ") || strings.Contains(string(data), "super-secret") {
		t.Fatalf("unsafe configuration bytes=%q err=%v", data, err)
	}
	second, changed, err := Create(root, false)
	if err != nil || changed || !second.Features.Orchestration {
		t.Fatalf("second=%#v changed=%t err=%v", second, changed, err)
	}
}

func TestApplyPreservesConfiguredPolicy(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	configuration, _, err := Create(root, false)
	if err != nil {
		t.Fatal(err)
	}
	configuration.Tools.AllowedPaths = []string{"internal/**"}
	data, err := localfile.EncodeJSON(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if err := localfile.AtomicReplace(Path(root), data, 0o644); err != nil {
		t.Fatal(err)
	}
	applied, changed, err := Apply(root)
	if err != nil || !changed || !applied.Features.Orchestration || !applied.Features.Sync || len(applied.Tools.AllowedPaths) != 1 || applied.Tools.AllowedPaths[0] != "internal/**" {
		t.Fatalf("applied=%+v changed=%t err=%v", applied, changed, err)
	}
}

func TestGatewayValidationRejectsRawOrUnsafeCredentials(t *testing.T) {
	configuration := Default()
	configuration.Providers = append(configuration.Providers, Provider{
		ID: "gateway", Kind: domain.ProviderKindOpenAIResponses, Enabled: true,
		Endpoint:   "https://gateway.example/v1/responses",
		Credential: Credential{Source: "env", Reference: "not-a-safe-name"},
		Models:     []Model{{ID: "model", Languages: []string{"*"}, ContextWindow: 100_000, SupportsTools: true, SupportsEditing: true, Efforts: []domain.ReasoningEffort{domain.EffortMedium}, CostClass: 2, LatencyClass: 2}},
	})
	if err := configuration.Validate(); err == nil {
		t.Fatal("unsafe credential reference was accepted")
	}
	configuration.Providers[len(configuration.Providers)-1].Credential.Reference = "L7_GATEWAY_KEY"
	if err := configuration.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestGatewayMayUseOnlyABoundedCatalog(t *testing.T) {
	configuration := Default()
	configuration.Providers = append(configuration.Providers, Provider{
		ID: "catalog-gateway", Kind: domain.ProviderKindOpenAIResponses, Enabled: true,
		Endpoint: "https://gateway.example/v1/responses", CatalogURL: "https://gateway.example/v1/l7-models",
		Credential: Credential{Source: "env", Reference: "L7_GATEWAY_KEY"}, Models: []Model{},
	})
	if err := configuration.Validate(); err != nil {
		t.Fatal(err)
	}
	configuration.Providers[len(configuration.Providers)-1].CatalogURL = ""
	if err := configuration.Validate(); err == nil {
		t.Fatal("gateway without configured models or a catalog was accepted")
	}
}

func TestStrictLoadRejectsUnknownAndDuplicateFields(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".l7"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{
		`{"schema":1,"schema":1}`,
		`{"schema":1,"unknown":true}`,
	} {
		if err := os.WriteFile(Path(root), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil {
			t.Fatalf("unsafe record was accepted: %s", data)
		}
	}
}

func FuzzStrictConfigurationDecode(f *testing.F) {
	valid, err := localfile.EncodeJSON(Default())
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Add([]byte(`{"schema":1,"schema":1}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > MaxBytes {
			t.Skip()
		}
		var configuration File
		if localfile.DecodeJSON(data, &configuration) == nil {
			_ = configuration.Validate()
		}
	})
}
