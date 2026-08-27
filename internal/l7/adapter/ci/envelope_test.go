package ci

import (
	"strings"
	"testing"

	"github.com/addressanup/level7-dev-loop/internal/l7/domain"
)

func TestDecodeAcceptsOneStrictReadyEnvelope(t *testing.T) {
	facts, err := Decode([]byte(validEnvelope()))
	if err != nil || !domain.EvaluateReadiness(facts).Ready || facts.Evidence.ChangeID != "product-change" {
		t.Fatalf("Decode()=%+v decision=%+v error=%v", facts, domain.EvaluateReadiness(facts), err)
	}
}

func TestDecodeRejectsMalformedOrExpandedEnvelope(t *testing.T) {
	tests := []string{
		`{"schema":1,"schema":1}`,
		strings.Replace(validEnvelope(), `"schema":1,`, `"schema":1,"unknown":true,`, 1),
		strings.Replace(validEnvelope(), `"owner":"accountable-owner",`, "", 1),
		validEnvelope() + `{}`,
		strings.Repeat("x", MaxEnvelopeBytes+1),
	}
	for _, data := range tests {
		if _, err := Decode([]byte(data)); err == nil {
			t.Fatalf("Decode(%q) unexpectedly passed", data[:min(len(data), 128)])
		}
	}
}

func TestDecodePreservesFalseReadyFactsForDeterministicDecision(t *testing.T) {
	data := strings.Replace(validEnvelope(), `"repository_clean":true`, `"repository_clean":false`, 1)
	facts, err := Decode([]byte(data))
	decision := domain.EvaluateReadiness(facts)
	if err != nil || decision.Ready || len(decision.Findings) == 0 {
		t.Fatalf("Decode()=%+v decision=%+v error=%v", facts, decision, err)
	}
}

func FuzzDecode(f *testing.F) {
	f.Add([]byte(validEnvelope()))
	f.Add([]byte(`{"schema":1,"schema":2}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		facts, err := Decode(data)
		if err == nil && len(data) > MaxEnvelopeBytes {
			t.Fatalf("oversized Decode()=%+v", facts)
		}
	})
}

func validEnvelope() string {
	return `{"schema":1,"change_id":"product-change","tier":3,"base_commit":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","candidate_commit":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","candidate_tree":"cccccccccccccccccccccccccccccccccccccccc","brief_commit":"dddddddddddddddddddddddddddddddddddddddd","configuration_digest":"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee","verification_commit":"ffffffffffffffffffffffffffffffffffffffff","review_commit":"1111111111111111111111111111111111111111","scope":["internal/product/**"],"checks":[{"name":"benchmark","benchmark":true,"passed":true,"exit_code":0,"code":"L7-VERIFY-000","message":"passed"}],"owner":"accountable-owner","implementer":"codex","reviewer":"claude","review_decision":"GO","benchmark_required":true,"plan_current":true,"repository_clean":true,"approval_current":true,"verification_current":true,"review_current":false,"audit_current":true}`
}
