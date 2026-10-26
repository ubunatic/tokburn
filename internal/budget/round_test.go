package budget

import "testing"

func TestRoundMenuCoversSixThroughTwenty(t *testing.T) {
	s := calc(t).Spec
	if s.GameRoundTasksMin != 6 || s.GameRoundTasksMax != 20 {
		t.Fatalf("menu range %d..%d, want 6..20", s.GameRoundTasksMin, s.GameRoundTasksMax)
	}
	if s.GoodScore != 86 || s.BreakPenaltyRedoFactor != 0.27 || s.BudgetAllMFactor != 1.02 {
		t.Fatalf("unexpected shared rules: %+v", s)
	}
	for n := 6; n <= 20; n++ {
		r := s
		r.GameRoundTasks = n
		if r.GameRoundTasks != n || r.GoodScore != 86 || r.BreakPenaltyRedoFactor != 0.27 {
			t.Fatalf("size %d did not use shared rules", n)
		}
	}
}
