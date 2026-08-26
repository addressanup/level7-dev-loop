package render

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

var sliceOneSourcePaths = []string{
	"schemas/semantic/guardrail.schema.json",
	"schemas/semantic/knowledge.schema.json",
	"schemas/semantic/obligation.schema.json",
	"schemas/semantic/taxonomy.schema.json",
	"semantic/taxonomy/guardrails.json",
	"semantic/taxonomy/knowledge.json",
	"semantic/taxonomy/obligations.json",
	"semantic/taxonomy/registry.json",
}

func TestDecodeRepositoryTaxonomyBundle(t *testing.T) {
	files := loadSliceOneSources(t)
	bundle, diagnostics := Decode(files)
	requireNoDiagnostics(t, diagnostics)
	if got, want := len(bundle.Taxonomies), 15; got != want {
		t.Fatalf("taxonomy count = %d, want %d", got, want)
	}
	if got, want := len(bundle.Obligations), 29; got != want {
		t.Fatalf("obligation count = %d, want %d", got, want)
	}
	if got, want := len(bundle.Guardrails), 12; got != want {
		t.Fatalf("guardrail count = %d, want %d", got, want)
	}
	if got, want := len(bundle.Knowledge), 5; got != want {
		t.Fatalf("knowledge count = %d, want %d", got, want)
	}
	if got, want := len(bundle.Descriptors), 4; got != want {
		t.Fatalf("descriptor count = %d, want %d", got, want)
	}
	if got, want := len(bundle.SourceDigests), len(files); got != want {
		t.Fatalf("source digest count = %d, want %d", got, want)
	}

	reversed := append([]SourceFile(nil), files...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	reordered, diagnostics := Decode(reversed)
	requireNoDiagnostics(t, diagnostics)
	if !reflect.DeepEqual(bundle, reordered) {
		t.Fatal("source-file ordering changed the decoded bundle")
	}
}

func TestDecodeFieldOrderMetamorphism(t *testing.T) {
	path := "schemas/semantic/taxonomy.schema.json"
	original := readRepositoryFile(t, path)
	first, diagnostics := Decode([]SourceFile{{Path: path, Data: original}})
	requireNoDiagnostics(t, diagnostics)

	var descriptor SchemaDescriptor
	if err := json.Unmarshal(original, &descriptor); err != nil {
		t.Fatal(err)
	}
	reordered, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	reordered = append(reordered, '\n')
	second, diagnostics := Decode([]SourceFile{{Path: path, Data: reordered}})
	requireNoDiagnostics(t, diagnostics)
	if !reflect.DeepEqual(first.Descriptors, second.Descriptors) {
		t.Fatal("JSON field order or insignificant whitespace changed typed semantics")
	}
	if first.SourceDigests[0].SHA256 == second.SourceDigests[0].SHA256 {
		t.Fatal("byte-distinct sources unexpectedly share a raw source digest")
	}
}

func TestDecodeCopiesMutableInputs(t *testing.T) {
	path := "schemas/semantic/taxonomy.schema.json"
	data := readRepositoryFile(t, path)
	bundle, diagnostics := Decode([]SourceFile{{Path: path, Data: data}})
	requireNoDiagnostics(t, diagnostics)
	before := bundle.Descriptors[0].ID
	for index := range data {
		data[index] = 'x'
	}
	if bundle.Descriptors[0].ID != before {
		t.Fatal("decoded value retained caller-owned source bytes")
	}
}

func TestSourceByteBounds(t *testing.T) {
	oversize := bytes.Repeat([]byte{' '}, MaxFileBytes+1)
	oversize[len(oversize)-1] = '\n'
	cases := []struct {
		name string
		data []byte
		rule string
	}{
		{name: "empty", data: nil, rule: "SEM-100"},
		{name: "oversize", data: oversize, rule: "SEM-100"},
		{name: "invalid UTF-8", data: []byte{0xff, '\n'}, rule: "SEM-103"},
		{name: "BOM", data: append([]byte{0xef, 0xbb, 0xbf}, []byte("{}\n")...), rule: "SEM-104"},
		{name: "missing final LF", data: []byte("{}"), rule: "SEM-105"},
		{name: "double final LF", data: []byte("{}\n\n"), rule: "SEM-105"},
		{name: "carriage return", data: []byte("{}\r\n"), rule: "SEM-105"},
		{name: "raw control", data: []byte{'{', 0x01, '}', '\n'}, rule: "SEM-106"},
		{name: "DEL", data: []byte{'{', 0x7f, '}', '\n'}, rule: "SEM-106"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			requireRule(t, validateSourceBytes("semantic/test.json", test.data), test.rule)
		})
	}
}

func TestStrictJSONShapeAndBounds(t *testing.T) {
	for depth := 0; depth < MaxJSONDepth; depth++ {
		payload := "{\"duplicate\":1,\"duplicate\":2}"
		for index := 0; index < depth; index++ {
			payload = "{\"nested\":" + payload + "}"
		}
		t.Run("duplicate-depth-"+string(rune('0'+depth)), func(t *testing.T) {
			requireRule(t, scanJSON("semantic/test.json", []byte(payload+"\n")), "SEM-110")
		})
	}
	maximumDepth := []byte(strings.Repeat("[", MaxJSONDepth) + "0" + strings.Repeat("]", MaxJSONDepth) + "\n")
	requireNoDiagnostics(t, scanJSON("semantic/test.json", maximumDepth))

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

	cases := []struct {
		name string
		data []byte
		rule string
	}{
		{name: "trailing value", data: []byte("{}\n{}\n"), rule: "SEM-108"},
		{name: "fraction", data: []byte("{\"n\":1.5}\n"), rule: "SEM-112"},
		{name: "exponent", data: []byte("{\"n\":1e2}\n"), rule: "SEM-112"},
		{name: "depth", data: []byte(strings.Repeat("[", MaxJSONDepth+1) + "0" + strings.Repeat("]", MaxJSONDepth+1) + "\n"), rule: "SEM-109"},
		{name: "array count", data: []byte("[" + strings.Repeat("0,", 2048) + "0]\n"), rule: "SEM-109"},
		{name: "object count", data: []byte(objectFields.String()), rule: "SEM-109"},
		{name: "string length", data: []byte("{\"v\":\"" + strings.Repeat("a", MaxStringBytes+1) + "\"}\n"), rule: "SEM-109"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			requireRule(t, scanJSON("semantic/test.json", test.data), test.rule)
		})
	}
}

func TestDecodeRejectsUnknownTrailingDuplicateAndUnsafeSources(t *testing.T) {
	cases := []struct {
		name  string
		files []SourceFile
		rule  string
	}{
		{
			name: "unknown typed field",
			files: []SourceFile{{
				Path: "semantic/taxonomy/registry.json",
				Data: []byte("{\"schema_version\":\"1.0.0\",\"taxonomies\":[],\"unknown\":true}\n"),
			}},
			rule: "SEM-113",
		},
		{
			name: "duplicate path",
			files: []SourceFile{
				{Path: "schemas/semantic/taxonomy.schema.json", Data: []byte("{}\n")},
				{Path: "schemas/semantic/taxonomy.schema.json", Data: []byte("{}\n")},
			},
			rule: "SEM-102",
		},
		{
			name:  "unsafe path",
			files: []SourceFile{{Path: "../semantic.json", Data: []byte("{}\n")}},
			rule:  "SEM-101",
		},
		{
			name:  "unsupported path",
			files: []SourceFile{{Path: "semantic/unknown.json", Data: []byte("{}\n")}},
			rule:  "SEM-101",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, diagnostics := Decode(test.files)
			requireRule(t, diagnostics, test.rule)
		})
	}
}

func TestDecodeAggregateBound(t *testing.T) {
	files := make([]SourceFile, 9)
	data := bytes.Repeat([]byte{' '}, MaxFileBytes)
	data[len(data)-1] = '\n'
	for index := range files {
		files[index] = SourceFile{
			Path: "schemas/semantic/bound-" + string(rune('a'+index)) + ".schema.json",
			Data: data,
		}
	}
	_, diagnostics := Decode(files)
	requireRule(t, diagnostics, "SEM-100")
}

func loadSliceOneBundle(t *testing.T) Bundle {
	t.Helper()
	bundle, diagnostics := Decode(loadSliceOneSources(t))
	requireNoDiagnostics(t, diagnostics)
	return bundle
}

func loadSliceOneSources(t *testing.T) []SourceFile {
	t.Helper()
	files := make([]SourceFile, 0, len(sliceOneSourcePaths))
	for _, path := range sliceOneSourcePaths {
		files = append(files, SourceFile{Path: path, Data: readRepositoryFile(t, path)})
	}
	return files
}

func readRepositoryFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
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
