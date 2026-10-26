// Package game is "Quota Defense": route a month of agent tasks to model tiers
// before the monthly budget burns down.
package game

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"ubunatic.com/tokburn/internal/budget"
)

// Card is one recurring agent task.
type Card struct {
	ID     string `yaml:"id" json:"id"`
	Task   string `yaml:"text" json:"text"`
	Need   string `yaml:"tier" json:"tier"`     // weakest tier that does the job
	Tokens int    `yaml:"tokens" json:"tokens"` // prompt size per run
	Runs   int    `yaml:"runs" json:"runs"`     // runs this month
}

// Deck is the validated embedded task pool; a round draws from it.
var Deck = mustLoadTasks()

var rank = map[string]int{"low": 0, "mid": 1, "high": 2}

// Result is the outcome of one routed card.
type Result struct {
	Card   Card
	Picked string
	USD    float64
	Broken bool
}

// Score routes one card: an underpowered tier breaks the task, so you pay
// for the failed run and the redo on the needed tier.
func Score(c budget.Calculator, card Card, picked string) (Result, error) {
	e, err := c.Estimate(card.Tokens, picked)
	if err != nil {
		return Result{}, err
	}
	r := Result{Card: card, Picked: picked, USD: e.USD * float64(card.Runs)}
	if rank[picked] < rank[card.Need] {
		redo, err := c.Estimate(card.Tokens, card.Need)
		if err != nil {
			return Result{}, err
		}
		r.Broken = true
		redoCost := redo.USD * float64(card.Runs)
		r.USD += redoCost + c.Spec.BreakPenaltyRedoFactor*redoCost
	}
	return r, nil
}

// DrawTasks shuffles the embedded deck until the round meets its tier minimums.
func DrawTasks(rng *rand.Rand, n, minPerTier int) []Card {
	if n > len(Deck) {
		n = len(Deck)
	}
	if n < 0 {
		n = 0
	}
	for {
		cards := append([]Card(nil), Deck...)
		rng.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
		draw := cards[:n]
		counts := map[string]int{}
		for _, card := range draw {
			counts[card.Need]++
		}
		if counts["low"] >= minPerTier && counts["mid"] >= minPerTier && counts["high"] >= minPerTier {
			return draw
		}
	}
}

// RoutingScore compares actual spending with the golden (right-tier) route.
// Perfect routing scores 100; unnecessary spending and failed runs reduce it.
func RoutingScore(goldenUSD, spentUSD float64) float64 {
	if spentUSD <= 0 {
		return 0
	}
	return 100 * goldenUSD / spentUSD
}

// Play runs one round of n cards, reading l/m/h choices from in.
func Play(c budget.Calculator, rng *rand.Rand, in io.Reader, out io.Writer) (float64, error) {
	cards := DrawTasks(rng, c.Spec.GameRoundTasks, c.Spec.MinPerTier)
	n := len(cards)
	budgetUSD := 0.0
	for _, card := range cards {
		r, err := Score(c, card, "mid")
		if err != nil {
			return 0, err
		}
		budgetUSD += r.USD
	}
	budgetUSD *= c.Spec.BudgetAllMFactor
	left := budgetUSD
	goldenUSD, spentUSD := 0.0, 0.0
	allMid := n > 0
	sc := bufio.NewScanner(in)
	fmt.Fprintf(out, "QUOTA DEFENSE: route %d tasks, keep your $%.2f round budget alive.\n", n, left)
	fmt.Fprintf(out, "Pick l(ow), m(id) or h(igh). Too weak = task breaks and is redone.\n\n")
	for i, card := range cards[:n] {
		fmt.Fprintf(out, "[%d/%d] %s  (%dk tokens x %d runs)\n> ", i+1, n, card.Task, card.Tokens/1000, card.Runs)
		picked := ""
		for picked == "" {
			if !sc.Scan() {
				return left, sc.Err()
			}
			picked = map[string]string{"l": "low", "m": "mid", "h": "high"}[strings.ToLower(strings.TrimSpace(sc.Text()))]
			if picked == "" {
				fmt.Fprint(out, "l, m or h? > ")
			}
		}
		r, err := Score(c, card, picked)
		if err != nil {
			return left, err
		}
		golden, err := Score(c, card, card.Need)
		if err != nil {
			return left, err
		}
		goldenUSD += golden.USD
		spentUSD += r.USD
		allMid = allMid && picked == "mid"
		left -= r.USD
		verdict := "ok"
		switch {
		case r.Broken:
			verdict = "BROKE, redone on " + card.Need
		case picked != card.Need:
			verdict = "works, but " + card.Need + " was enough"
		}
		fmt.Fprintf(out, "  %s: -$%.2f  %s  $%.2f left\n\n", verdict, r.USD, budget.Bar(budgetUSD-left, budgetUSD, 24), left)
	}
	if left > 0 {
		fmt.Fprintf(out, "Month survived with $%.2f left. Score: %.2f\n", left, RoutingScore(goldenUSD, spentUSD))
		fmt.Fprintf(out, "Good score: %.0f or higher (perfect routing: 100).\n", c.Spec.GoodScore)
	} else {
		fmt.Fprintf(out, "Quota on fire: $%.2f over budget. Score: %.2f\n", -left, RoutingScore(goldenUSD, spentUSD))
		fmt.Fprintf(out, "Good score: %.0f or higher (perfect routing: 100).\n", c.Spec.GoodScore)
	}
	if allMid {
		fmt.Fprintln(out, "Easter egg: all-m — middle management keeps the month alive.")
	}
	return left, nil
}
