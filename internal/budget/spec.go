package budget

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Tier is one model price class from spec/pricing.yaml.
type Tier struct {
	Name          string  `yaml:"name"`
	USDPerMillion float64 `yaml:"usd_per_million"`
	GoodFor       string  `yaml:"good_for"`
}

// Spec mirrors spec/pricing.yaml.
type Spec struct {
	CliffTokens         int     `yaml:"cliff_tokens"`
	CliffInclusive      bool    `yaml:"cliff_inclusive"`
	SurchargeMultiplier float64 `yaml:"surcharge_multiplier"`
	MonthlyBudgetUSD    float64 `yaml:"monthly_budget_usd"`
	Tiers               []Tier  `yaml:"tiers"`
}

// LoadSpec parses and validates a pricing spec.
func LoadSpec(data []byte) (Spec, error) {
	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse pricing spec: %w", err)
	}
	if s.CliffTokens <= 0 || s.SurchargeMultiplier < 1 || len(s.Tiers) == 0 {
		return s, fmt.Errorf("invalid pricing spec: need cliff_tokens > 0, surcharge_multiplier >= 1, tiers")
	}
	return s, nil
}

// Tier returns the tier with the given name.
func (s Spec) Tier(name string) (Tier, error) {
	for _, t := range s.Tiers {
		if t.Name == name {
			return t, nil
		}
	}
	return Tier{}, fmt.Errorf("unknown tier %q", name)
}
