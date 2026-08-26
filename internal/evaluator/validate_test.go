package evaluator

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/render"
)

func TestRepositoryControlsDecodeValidateAndBindFinalInputs(t *testing.T) {
	controls := loadRepositoryControls(t)
	if len(controls.Cases) != 46 || len(controls.TruthLabels) != 46 || len(controls.Graders) != 17 || len(controls.Coverage.Axes) != 29 || len(controls.Descriptors) != 7 || len(controls.SourceDigests) != 13 || len(controls.Bindings) != 6 {
		t.Fatalf("control shape: cases=%d truths=%d graders=%d axes=%d descriptors=%d sources=%d bindings=%d", len(controls.Cases), len(controls.TruthLabels), len(controls.Graders), len(controls.Coverage.Axes), len(controls.Descriptors), len(controls.SourceDigests), len(controls.Bindings))
	}
	requireNoDiagnostics(t, ValidateInputBindings(
		controls,
		readRepositoryFile(t, "fixtures/public/bl-002/semantic-cases.json"),
		readRepositoryFile(t, "fixtures/public/bl-002/broken-candidates.json"),
	))
	if controls.Protocol.ID != PublicProtocolID || controls.Protocol.Version != PublicProtocolVersion || controls.Protocol.RunCount != 2 || controls.Protocol.SeedPolicy.Seed != 0 || controls.Protocol.HostModelPolicy.Host != "not_applicable" || controls.Protocol.HostModelPolicy.Model != "not_applicable" {
		t.Fatalf("public protocol identity or deterministic values differ: %+v", controls.Protocol)
	}
	truthByCase := make(map[string]TruthLabel, len(controls.TruthLabels))
	for _, truth := range controls.TruthLabels {
		truthByCase[truth.CaseID] = truth
	}
	var semanticFixtures struct {
		Cases []struct {
			ID               string `json:"id"`
			ExpectedDecision string `json:"expected_decision"`
			ExpectedRuleID   string `json:"expected_rule_id"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepositoryFile(t, "fixtures/public/bl-002/semantic-cases.json"), &semanticFixtures); err != nil {
		t.Fatal(err)
	}
	if len(semanticFixtures.Cases) != 9 {
		t.Fatalf("semantic fixture count = %d, want 9", len(semanticFixtures.Cases))
	}
	for _, fixture := range semanticFixtures.Cases {
		truth := truthByCase[fixture.ID]
		expectedRules := []string{}
		if fixture.ExpectedRuleID != "" {
			expectedRules = []string{fixture.ExpectedRuleID}
		}
		if truth.ExpectedDecision != fixture.ExpectedDecision || !equalStrings(truth.ExpectedRuleIDs, expectedRules) {
			t.Fatalf("semantic fixture truth mismatch: fixture=%+v truth=%+v", fixture, truth)
		}
	}
}

func TestDecodeControlsIsOrderIndependentAndCopiesInput(t *testing.T) {
	files := loadRepositoryControlFiles(t)
	first, diagnostics := DecodeControls(files)
	requireNoDiagnostics(t, diagnostics)
	reversed := append([]ControlFile(nil), files...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	second, diagnostics := DecodeControls(reversed)
	requireNoDiagnostics(t, diagnostics)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("control file ordering changed decoded evaluator semantics")
	}
	before := first.Protocol.ID
	for index := range files[0].Data {
		files[0].Data[index] = 'x'
	}
	if first.Protocol.ID != before {
		t.Fatal("decoded controls retained caller-owned bytes")
	}
}

func TestControlFieldOrderMetamorphismPreservesTypedSemantics(t *testing.T) {
	files := loadRepositoryControlFiles(t)
	for index := range files {
		if files[index].Path != "fixtures/public/bl-003/protocol.json" {
			continue
		}
		var document any
		if err := json.Unmarshal(files[index].Data, &document); err != nil {
			t.Fatal(err)
		}
		reordered, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		files[index].Data = append(reordered, '\n')
	}
	actual, diagnostics := DecodeControls(files)
	requireNoDiagnostics(t, diagnostics)
	expected := loadRepositoryControls(t)
	if !reflect.DeepEqual(actual.Protocol, expected.Protocol) || !reflect.DeepEqual(actual.Cases, expected.Cases) || actual.SourceSetSHA256 == expected.SourceSetSHA256 {
		t.Fatal("field ordering changed typed semantics or failed to change raw source identity")
	}
}

func TestStrictControlBytesJSONAndUnknownFieldsFailClosed(t *testing.T) {
	oversize := bytes.Repeat([]byte{' '}, MaxFileBytes+1)
	oversize[len(oversize)-1] = '\n'
	byteCases := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: nil},
		{name: "oversize", data: oversize},
		{name: "invalid UTF-8", data: []byte{0xff, '\n'}},
		{name: "BOM", data: append([]byte{0xef, 0xbb, 0xbf}, []byte("{}\n")...)},
		{name: "missing LF", data: []byte("{}")},
		{name: "double LF", data: []byte("{}\n\n")},
		{name: "carriage return", data: []byte("{}\r\n")},
		{name: "terminal control", data: []byte{'{', 0x1b, '}', '\n'}},
	}
	for _, test := range byteCases {
		t.Run(test.name, func(t *testing.T) {
			requireRule(t, validateControlBytes("fixtures/public/bl-003/protocol.json", test.data), "EVAL-200")
		})
	}

	for depth := 0; depth < MaxJSONDepth; depth++ {
		payload := "{\"duplicate\":1,\"duplicate\":2}"
		for index := 0; index < depth; index++ {
			payload = "{\"nested\":" + payload + "}"
		}
		requireRule(t, scanJSON("control.json", []byte(payload+"\n")), "EVAL-201")
	}
	var objectFields strings.Builder
	objectFields.WriteByte('{')
	for index := 0; index <= MaxObjectFields; index++ {
		if index != 0 {
			objectFields.WriteByte(',')
		}
		objectFields.WriteString("\"f")
		objectFields.WriteString(strconv.Itoa(index))
		objectFields.WriteString("\":0")
	}
	objectFields.WriteString("}\n")
	for _, data := range [][]byte{
		[]byte("{}\n{}\n"),
		[]byte("{\"n\":1.5}\n"),
		[]byte("{\"n\":1e2}\n"),
		[]byte(strings.Repeat("[", MaxJSONDepth+1) + "0" + strings.Repeat("]", MaxJSONDepth+1) + "\n"),
		[]byte("[" + strings.Repeat("0,", MaxCoverageLinks) + "0]\n"),
		[]byte(objectFields.String()),
		[]byte("{\"v\":\"" + strings.Repeat("a", MaxStringBytes+1) + "\"}\n"),
	} {
		requireRule(t, scanJSON("control.json", data), "EVAL-200")
	}

	files := loadRepositoryControlFiles(t)
	for index := range files {
		if files[index].Path == "fixtures/public/bl-003/protocol.json" {
			files[index].Data = bytes.Replace(files[index].Data, []byte("\"schema_version\": \"1.0.0\","), []byte("\"schema_version\": \"1.0.0\",\n  \"unknown\": true,"), 1)
		}
	}
	_, diagnostics := DecodeControls(files)
	requireRule(t, diagnostics, "EVAL-202")

	extra := append(loadRepositoryControlFiles(t), ControlFile{Path: "schemas/evaluation/extra.schema.json", Data: []byte("{}\n")})
	_, diagnostics = DecodeControls(extra)
	requireRule(t, diagnostics, "EVAL-200")
}

func TestExactCaseRosterAndDeterministicSupportFailClosed(t *testing.T) {
	controls := loadRepositoryControls(t)
	for index := range controls.Cases {
		if controls.Cases[index].ID == "L7-CASE-BL002-BOUNDARY" {
			controls.Cases[index].ID = "L7-CASE-BL002-BOUNDARY-X"
			break
		}
	}
	requireRule(t, validateCasesAndTruth(controls.Cases, controls.TruthLabels, controls.Graders), "EVAL-213")

	controls = loadRepositoryControls(t)
	controls.Cases[0].GraderIDs = []string{"L7-EGR-MODEL-JUDGE-SUPPLEMENTAL"}
	requireRule(t, validateCasesAndTruth(controls.Cases, controls.TruthLabels, controls.Graders), "EVAL-215")

	controls = loadRepositoryControls(t)
	for index := range controls.Cases {
		if controls.Cases[index].ID == "L7-CASE-BL002-BROKEN-CANARY" {
			controls.Cases[index].GraderIDs = []string{"L7-EGR-SEMANTIC-GUARDRAILS"}
			break
		}
	}
	requireRule(t, validateCasesAndTruth(controls.Cases, controls.TruthLabels, controls.Graders), "EVAL-215")

	controls = loadRepositoryControls(t)
	for index := range controls.Cases {
		if controls.Cases[index].ID == "L7-CASE-BL002-BROKEN-CANARY" {
			controls.Cases[index].InputFixture = "fixtures/public/bl-002/semantic-cases.json"
			controls.Cases[index].InputDigest = SemanticCasesSHA256
			break
		}
	}
	requireRule(t, validateCasesAndTruth(controls.Cases, controls.TruthLabels, controls.Graders), "EVAL-214")

	controls = loadRepositoryControls(t)
	controls.Graders[0].ObligationIDs = []string{"L7-OBL-UNKNOWN-999"}
	requireRule(t, validateGraders(controls.Graders, controls.Cases, controls.TruthLabels), "EVAL-222")
}

func TestDiagnosticsRedactCanariesAndRespectAggregateBound(t *testing.T) {
	const canary = "L7_SYNTHETIC_CANARY_SUPERSECRET123"
	diagnostic := newDiagnostic("EVAL-200", canary, "unknown field "+canary, "remove "+canary)
	encoded, err := json.Marshal(diagnostic)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte("L7_SYNTHETIC_CANARY_")) || bytes.Contains(encoded, []byte("SUPERSECRET123")) {
		t.Fatalf("diagnostic retained a synthetic canary: %s", encoded)
	}

	large := Diagnostic{
		Rule:    strings.Repeat("r", 32),
		Subject: strings.Repeat("s", 160),
		Message: strings.Repeat("m", MaxDiagnosticBytes),
		Next:    strings.Repeat("n", MaxDiagnosticBytes),
	}
	var diagnostics []Diagnostic
	for index := 0; index < MaxDiagnostics; index++ {
		diagnostics = appendDiagnostics(diagnostics, large)
	}
	total := 0
	for _, item := range diagnostics {
		total += len(item.Rule) + len(item.Subject) + len(item.Message) + len(item.Next)
	}
	if total > MaxDiagnosticsBytes {
		t.Fatalf("diagnostic bytes = %d, want at most %d", total, MaxDiagnosticsBytes)
	}

	forwardInput := make([]Diagnostic, MaxDiagnostics+16)
	for index := range forwardInput {
		forwardInput[index] = newDiagnostic("EVAL-200", "subject-"+strconv.Itoa(index), "bounded failure", "supply bounded input")
	}
	reverseInput := append([]Diagnostic(nil), forwardInput...)
	for left, right := 0, len(reverseInput)-1; left < right; left, right = left+1, right-1 {
		reverseInput[left], reverseInput[right] = reverseInput[right], reverseInput[left]
	}
	forward := finishDiagnostics(appendDiagnostics(nil, forwardInput...))
	reverse := finishDiagnostics(appendDiagnostics(nil, reverseInput...))
	if !reflect.DeepEqual(forward, reverse) {
		t.Fatal("bounded diagnostic selection depends on discovery order")
	}
}

func TestProtocolTruthGraderAdjudicationAndHoldoutMutationsFail(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Controls)
		rule   string
	}{
		{name: "protocol identity", rule: "EVAL-206", mutate: func(controls *Controls) { controls.Protocol.Version = "2.0.0" }},
		{name: "protocol trial count", rule: "EVAL-210", mutate: func(controls *Controls) { controls.Protocol.RunCount = 1 }},
		{name: "case truth mismatch", rule: "EVAL-218", mutate: func(controls *Controls) { controls.TruthLabels[0].CaseID = "L7-CASE-UNKNOWN" }},
		{name: "truth changed by candidate", rule: "EVAL-219", mutate: func(controls *Controls) { controls.TruthLabels[0].ExpectedDecision = "blocked" }},
		{name: "deterministic grader becomes model", rule: "EVAL-224", mutate: func(controls *Controls) { controls.Graders[0].Class = "model" }},
		{name: "model judge self adjudicates", rule: "EVAL-225", mutate: func(controls *Controls) { controls.Graders[6].AuthorityLimit = "may-change-truth" }},
		{name: "candidate adjudicator", rule: "EVAL-226", mutate: func(controls *Controls) { controls.Adjudication.EligibleRole = "candidate" }},
		{name: "holdout below twenty percent", rule: "EVAL-270", mutate: func(controls *Controls) { controls.Protocol.ProtectedHoldout.MinimumCorpusBasisPoints = 1999 }},
		{name: "holdout instantiated", rule: "EVAL-270", mutate: func(controls *Controls) { controls.Protocol.ProtectedHoldout.ImplementationState = "active" }},
		{name: "control roster changed", rule: "EVAL-269", mutate: func(controls *Controls) { controls.Protocol.ControlPaths = controls.Protocol.ControlPaths[1:] }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controls := loadRepositoryControls(t)
			test.mutate(&controls)
			requireRule(t, ValidateControls(controls), test.rule)
		})
	}
}

func TestSchemaDescriptorParityAndRunManifestBounds(t *testing.T) {
	controls := loadRepositoryControls(t)
	foundRunManifest := false
	for _, descriptor := range controls.Descriptors {
		if descriptor.ID != "schemas/evaluation/run-manifest.schema.json" {
			continue
		}
		foundRunManifest = true
		if descriptor.Properties["candidate"].AdditionalProperties == nil || *descriptor.Properties["candidate"].AdditionalProperties || descriptor.Properties["trials"].MaxItems == nil || *descriptor.Properties["trials"].MaxItems != MaxTrials || descriptor.Properties["latency"].Maximum == nil || *descriptor.Properties["latency"].Maximum != 5000 || descriptor.Properties["cost"].Maximum == nil || *descriptor.Properties["cost"].Maximum != 0 {
			t.Fatal("run-manifest descriptor lost closed identity, trial, cost, or latency bounds")
		}
	}
	if !foundRunManifest {
		t.Fatal("run-manifest descriptor missing")
	}
	delete(controls.Descriptors[0].Properties, "trigger")
	requireRule(t, ValidateControls(controls), "EVAL-207")

	controls = loadRepositoryControls(t)
	for index := range controls.Descriptors {
		if controls.Descriptors[index].ID != "schemas/evaluation/run-manifest.schema.json" {
			continue
		}
		candidate := controls.Descriptors[index].Properties["candidate"]
		candidate.Required = []string{"path", "path"}
		controls.Descriptors[index].Properties["candidate"] = candidate
	}
	requireRule(t, validateDescriptors(controls.Descriptors), "EVAL-209")
}

func TestInputDigestBindingRejectsEitherChangedBL002Input(t *testing.T) {
	controls := loadRepositoryControls(t)
	semanticCases := readRepositoryFile(t, "fixtures/public/bl-002/semantic-cases.json")
	brokenCandidates := readRepositoryFile(t, "fixtures/public/bl-002/broken-candidates.json")
	mutatedSemantic := append([]byte(nil), semanticCases...)
	mutatedSemantic[0] = '['
	requireRule(t, ValidateInputBindings(controls, mutatedSemantic, brokenCandidates), "EVAL-268")
	mutatedBroken := append([]byte(nil), brokenCandidates...)
	mutatedBroken[0] = '['
	requireRule(t, ValidateInputBindings(controls, semanticCases, mutatedBroken), "EVAL-268")
	controls.Bindings[0].Inputs.SemanticCases.SHA256 = strings.Repeat("0", 64)
	requireRule(t, ValidateControls(controls), "EVAL-268")

	oversize := bytes.Repeat([]byte{'x'}, MaxFileBytes+1)
	requireRule(t, ValidateInputBindings(controls, oversize, brokenCandidates), "EVAL-200")
}

func TestControlManifestBindingRejectsRowsOrderAndDigestDrift(t *testing.T) {
	controls, diagnostics := DecodeControls(loadRepositoryControlFiles(t))
	requireNoDiagnostics(t, diagnostics)
	manifest := readRepositoryFile(t, "harness/wave-02-evaluator-controls.sha256")
	bound, diagnostics := BindControlManifest(controls, manifest)
	requireNoDiagnostics(t, diagnostics)
	requireNoDiagnostics(t, ValidateControls(bound))

	missing := manifest[:bytes.LastIndexByte(manifest[:len(manifest)-1], '\n')+1]
	_, diagnostics = BindControlManifest(controls, missing)
	requireRule(t, diagnostics, "EVAL-266")

	tampered := append([]byte(nil), manifest...)
	tampered[0] = 'f'
	_, diagnostics = BindControlManifest(controls, tampered)
	requireRule(t, diagnostics, "EVAL-266")

	lines := strings.Split(strings.TrimSuffix(string(manifest), "\n"), "\n")
	lines[0], lines[1] = lines[1], lines[0]
	_, diagnostics = BindControlManifest(controls, []byte(strings.Join(lines, "\n")+"\n"))
	requireRule(t, diagnostics, "EVAL-266")
}

func TestTypedControlGraphRejectsCallerExpansionBeforeCanonicalHashing(t *testing.T) {
	controls := loadRepositoryControls(t)
	controls.Cases = make([]Case, MaxCoverageLinks+1)
	diagnostics := ValidateControls(controls)
	requireRule(t, diagnostics, "EVAL-200")
	if len(diagnostics) != 1 {
		t.Fatalf("unbounded typed graph was traversed after rejection: %+v", diagnostics)
	}

	controls = loadRepositoryControls(t)
	controls.SourceDigests[0].SHA256 = strings.Repeat("f", 64)
	controls.SourceSetSHA256 = sourceSetSHA256(controls.SourceDigests)
	requireRule(t, ValidateControls(controls), "EVAL-265")
}

func loadRepositoryControls(t *testing.T) Controls {
	t.Helper()
	controls, diagnostics := DecodeControls(loadRepositoryControlFiles(t))
	requireNoDiagnostics(t, diagnostics)
	controls, diagnostics = BindControlManifest(controls, readRepositoryFile(t, "harness/wave-02-evaluator-controls.sha256"))
	requireNoDiagnostics(t, diagnostics)
	requireNoDiagnostics(t, ValidateControls(controls))
	return controls
}

func loadRepositoryControlFiles(t *testing.T) []ControlFile {
	t.Helper()
	files := make([]ControlFile, 0, len(expectedJSONControlPaths))
	for _, relative := range expectedJSONControlPaths {
		files = append(files, ControlFile{Path: relative, Data: readRepositoryFile(t, relative)})
	}
	return files
}

func readRepositoryFile(t *testing.T, relative string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(relative)))
	if err != nil {
		t.Fatalf("read %s: %v", relative, err)
	}
	return data
}

func requireNoDiagnostics(t *testing.T, diagnostics []Diagnostic) {
	t.Helper()
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
}

func requireRule(t *testing.T, diagnostics []Diagnostic, rule string) {
	t.Helper()
	for _, diagnostic := range diagnostics {
		if diagnostic.Rule == rule {
			return
		}
	}
	t.Fatalf("missing %s in diagnostics: %+v", rule, diagnostics)
}

func descriptorByID(t *testing.T, descriptors []render.SchemaDescriptor, id string) render.SchemaDescriptor {
	t.Helper()
	for _, descriptor := range descriptors {
		if descriptor.ID == id {
			return descriptor
		}
	}
	t.Fatalf("descriptor %s not found", id)
	return render.SchemaDescriptor{}
}

func sortedDiagnosticKeys(diagnostics []Diagnostic) []string {
	keys := make([]string, len(diagnostics))
	for index, diagnostic := range diagnostics {
		keys[index] = diagnostic.Rule + "\x00" + diagnostic.Subject + "\x00" + diagnostic.Message
	}
	sort.Strings(keys)
	return keys
}
