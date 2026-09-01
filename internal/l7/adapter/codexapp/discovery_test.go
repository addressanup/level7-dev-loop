package codexapp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDiscoverSequencesAppServerHandshake(t *testing.T) {
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	helper := filepath.Join(directory, "codex-fixture")
	script := `#!/bin/sh
read_request() {
  IFS= read -r request || exit 91
  case "$request" in *"$1"*) ;; *) exit 92 ;; esac
}
read_request '"method":"initialize"'
printf '%s\n' '{"id":1,"result":{"serverInfo":{"name":"fixture"}}}'
read_request '"method":"initialized"'
read_request '"method":"account/read"'
printf '%s\n' '{"id":2,"result":{"account":{"type":"chatgpt"}}}'
read_request '"method":"model/list"'
printf '%s\n' '{"id":3,"result":{"data":[{"id":"fixture-model"}]}}'
read_request '"method":"account/rateLimits/read"'
printf '%s\n' '{"id":4,"result":{"rateLimits":{}}}'
`
	if err := os.WriteFile(helper, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	output, err := Discover(ctx, helper)
	if err != nil {
		t.Fatal(err)
	}
	if text := string(output); !strings.Contains(text, `"id":2`) || !strings.Contains(text, `fixture-model`) || !strings.Contains(text, `"id":4`) {
		t.Fatalf("discovery output=%q", text)
	}
}
