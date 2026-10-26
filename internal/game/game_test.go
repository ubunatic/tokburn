package game

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"

	"ubunatic.com/tokburn/internal/budget"
	"ubunatic.com/tokburn/spec"
)

func calc(t *testing.T) budget.Calculator {
	t.Helper()
	s, err := budget.LoadSpec(spec.Pricing)
	if err != nil {
		t.Fatal(err)
	}
	return budget.Calculator{Spec: s}
}

func TestScoreUnderpoweredBreaks(t *testing.T) {
	c := calc(t)
	card := Card{"x", "high", 10_000, 1}
	weak, err := Score(c, card, "low")
	if err != nil {
		t.Fatal(err)
	}
	right, _ := Score(c, card, "high")
	if !weak.Broken || weak.USD <= right.USD {
		t.Fatalf("underpowered pick must break and cost more: weak=%+v right=%+v", weak, right)
	}
}

func TestPlayRound(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("x\nh\nh\nh\n")
	left, err := Play(calc(t), rand.New(rand.NewSource(1)), 3, in, &out)
	if err != nil {
		t.Fatal(err)
	}
	if left >= 30 || !strings.Contains(out.String(), "l, m or h?") {
		t.Fatalf("left=%v output:\n%s", left, out.String())
	}
}
