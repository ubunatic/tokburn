package budget

import (
	"math"
	"testing"

	"ubunatic.com/tokburn/spec"
)

func calc(t *testing.T) Calculator {
	t.Helper()
	s, err := LoadSpec(spec.Pricing)
	if err != nil {
		t.Fatal(err)
	}
	return Calculator{Spec: s}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestEstimateBase(t *testing.T) {
	e, err := calc(t).Estimate(50_000, "mid")
	if err != nil {
		t.Fatal(err)
	}
	if e.Surcharged || !near(e.USD, 0.075) {
		t.Fatalf("got %+v, want 0.075 USD without surcharge", e)
	}
}

func TestEstimateUnknownTier(t *testing.T) {
	if _, err := calc(t).Estimate(1, "ultra"); err == nil {
		t.Fatal("want error for unknown tier")
	}
}

func TestCountTokens(t *testing.T) {
	if got := CountTokens([]byte("abcdefgh")); got != 2 {
		t.Fatalf("got %d, want 2", got)
	}
	if got := CountTokens([]byte("abcde")); got != 2 {
		t.Fatalf("got %d, want 2 (round up)", got)
	}
}
