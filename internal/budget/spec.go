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
	TitleFrameSeconds      float64 `yaml:"title_frame_seconds"`
	CliffTokens            int     `yaml:"cliff_tokens"`
	CliffInclusive         bool    `yaml:"cliff_inclusive"`
	SurchargeMultiplier    float64 `yaml:"surcharge_multiplier"`
	GameRoundTasks         int     `yaml:"game_round_tasks"`
	GameRoundTasksMin      int     `yaml:"game_round_tasks_min"`
	GameRoundTasksMax      int     `yaml:"game_round_tasks_max"`
	GameRoundTasksShort    int     `yaml:"game_round_tasks_short"`
	MinPerTier             int     `yaml:"min_per_tier"`
	BudgetAllMFactor       float64 `yaml:"budget_all_m_factor"`
	BreakPenaltyRedoFactor float64 `yaml:"break_penalty_redo_factor"`
	GoodScore              float64 `yaml:"good_score"`
	JokerStart             int     `yaml:"joker_start"`
	JokerComboAnswers      int     `yaml:"joker_combo_answers"`
	JokerFastSeconds       float64 `yaml:"joker_fast_seconds"`
	JokerComboReward       int     `yaml:"joker_combo_reward"`
	Tiers                  []Tier  `yaml:"tiers"`
}

// LoadSpec parses and validates a pricing spec.
func LoadSpec(data []byte) (Spec, error) {
	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse pricing spec: %w", err)
	}
	if !(s.TitleFrameSeconds > 0 && s.TitleFrameSeconds <= 60) {
		return s, fmt.Errorf("invalid pricing spec: title_frame_seconds must be in (0, 60]")
	}
	if s.JokerStart < 0 || s.JokerComboAnswers <= 0 || s.JokerFastSeconds <= 0 || s.JokerComboReward <= 0 {
		return s, fmt.Errorf("invalid pricing spec: joker_start >= 0, joker_combo_answers, joker_fast_seconds and joker_combo_reward > 0 required")
	}
	if s.GameRoundTasks <= 0 || s.MinPerTier < 0 || s.MinPerTier*3 > s.GameRoundTasks || s.BudgetAllMFactor <= 0 || s.BreakPenaltyRedoFactor < 0 || s.GoodScore <= 0 || s.GoodScore > 100 || s.CliffTokens <= 0 || s.SurchargeMultiplier < 1 || len(s.Tiers) == 0 {
		return s, fmt.Errorf("invalid pricing spec: need game_round_tasks > 0, min_per_tier >= 0 and 3*min_per_tier <= game_round_tasks, budget_all_m_factor > 0, break_penalty_redo_factor >= 0, good_score in (0, 100], cliff_tokens > 0, surcharge_multiplier >= 1, tiers")
	}
	if s.GameRoundTasksMin < 3*s.MinPerTier || s.GameRoundTasksMin <= 0 || s.GameRoundTasksMax < s.GameRoundTasksMin || s.GameRoundTasksShort < s.GameRoundTasksMin || s.GameRoundTasksShort > s.GameRoundTasksMax {
		return s, fmt.Errorf("invalid pricing spec: menu round limits must contain the short default and meet tier minimums")
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

// ShortRound returns a copy with the configured short-mode size active.
func (s Spec) ShortRound() Spec {
	s.GameRoundTasks = s.GameRoundTasksShort
	return s
}
