package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfigurationIsValidAndFeatureOff(t *testing.T) {
	configuration := Default(false)
	if err := configuration.Validate(); err != nil {
		t.Fatal(err)
	}
	if configuration.Features.LocalLifecycle || configuration.Domain().LocalLifecycle {
		t.Fatal("local lifecycle is not default OFF")
	}
}

func TestAdoptCreatesIdempotentlyAndEnablesOnlyWhenExplicit(t *testing.T) {
	root := physicalRoot(t)
	configuration, changed, err := Adopt(root, false)
	if err != nil || !changed || configuration.Features.LocalLifecycle {
		t.Fatalf("first Adopt() config=%+v changed=%v error=%v", configuration, changed, err)
	}
	_, changed, err = Adopt(root, false)
	if err != nil || changed {
		t.Fatalf("idempotent Adopt() changed=%v error=%v", changed, err)
	}
	configuration, changed, err = Adopt(root, true)
	if err != nil || !changed || !configuration.Features.LocalLifecycle {
		t.Fatalf("enabled Adopt() config=%+v changed=%v error=%v", configuration, changed, err)
	}
	loaded, err := Load(root)
	if err != nil || !loaded.Features.LocalLifecycle {
		t.Fatalf("Load() config=%+v error=%v", loaded, err)
	}
}

func TestLoadRejectsMalformedConfigurationWithoutRepair(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"duplicate", `{"schema":1,"schema":1}`},
		{"unknown", `{"schema":1,"unknown":true}`},
		{"trailing", `{"schema":1} {}`},
		{"future", `{"schema":2,"features":{"local_lifecycle":false},"verification":[],"limits":{"max_input_bytes":1048576,"max_git_output_bytes":16777216,"max_git_paths":100000,"max_command_output_bytes":8388608,"max_command_seconds":1800},"protected_paths":[],"providers":{"implementer":"","reviewer":""}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := physicalRoot(t)
			directory := filepath.Join(root, ".l7")
			if err := os.Mkdir(directory, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(directory, "config.json")
			if err := os.WriteFile(path, []byte(test.data), 0o644); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, _, err := Adopt(root, true); err == nil {
				t.Fatal("Adopt() repaired malformed configuration without permission")
			}
			after, err := os.ReadFile(path)
			if err != nil || string(after) != string(before) {
				t.Fatalf("malformed configuration changed: before=%q after=%q error=%v", before, after, err)
			}
		})
	}
}

func TestLoadRejectsUnsafeConfigurationFile(t *testing.T) {
	root := physicalRoot(t)
	directory := filepath.Join(root, ".l7")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(directory, "config.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("symlink Load() error=%v", err)
	}
}

func TestValidateRejectsUnsafeCommandsPathsProvidersAndLimits(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*File)
	}{
		{"empty argv", func(file *File) { file.Verification = []VerificationCommand{{Name: "test"}} }},
		{"duplicate command", func(file *File) {
			file.Verification = []VerificationCommand{{Name: "test", Argv: []string{"make"}}, {Name: "test", Argv: []string{"go", "test"}}}
		}},
		{"unsafe path", func(file *File) { file.ProtectedPaths = []string{"../outside"} }},
		{"duplicate path", func(file *File) { file.ProtectedPaths = []string{"safe/**", "safe/**"} }},
		{"provider", func(file *File) { file.Providers.Reviewer = "shell" }},
		{"limit", func(file *File) { file.Limits.MaxGitPaths = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configuration := Default(false)
			test.mutate(&configuration)
			if err := configuration.Validate(); err == nil {
				t.Fatal("Validate() unexpectedly passed")
			}
		})
	}
}

func TestValidateRepositoryPath(t *testing.T) {
	for _, value := range []string{"README.md", "internal/l7/**", "space allowed/file.go", ".github/workflows/ci.yml"} {
		if err := ValidateRepositoryPath(value); err != nil {
			t.Fatalf("ValidateRepositoryPath(%q)=%v", value, err)
		}
	}
	for _, value := range []string{"", "/absolute", "../escape", "a/../b", `a\\b`, "wild*card", "directory/"} {
		if err := ValidateRepositoryPath(value); err == nil {
			t.Fatalf("ValidateRepositoryPath(%q) unexpectedly passed", value)
		}
	}
}

func physicalRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}
