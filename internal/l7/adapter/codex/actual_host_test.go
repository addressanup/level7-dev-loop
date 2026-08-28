//go:build l7_actual_provider

package codex

import (
	"context"
	"os"
	"testing"
)

func TestActualHostProbe(t *testing.T) {
	if os.Getenv("L7_AUTHORIZE_ACTUAL_CODEX") != "probe" {
		t.Skip("set L7_AUTHORIZE_ACTUAL_CODEX=probe only under a separately approved actual-host envelope")
	}
	identity, err := New().Probe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("provider=%s version=%q executable=%q digest=%s capability=%s", identity.Provider, identity.Version, identity.Executable, identity.Digest, identity.Capability)
}
