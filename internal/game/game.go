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
		r.USD += redo.USD * float64(card.Runs)
	}
	return r, nil
}

// Play runs one round of n cards, reading l/m/h choices from in.
func Play(c budget.Calculator, rng *rand.Rand, n int, in io.Reader, out io.Writer) (float64, error) {
	cards := append([]Card(nil), Deck...)
	rng.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	if n > len(cards) {
		n = len(cards)
	}
	left := c.Spec.MonthlyBudgetUSD
	sc := bufio.NewScanner(in)
	fmt.Fprintf(out, "QUOTA DEFENSE: route %d tasks, keep your $%.0f/month alive.\n", n, left)
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
		left -= r.USD
		verdict := "ok"
		switch {
		case r.Broken:
			verdict = "BROKE, redone on " + card.Need
		case picked != card.Need:
			verdict = "works, but " + card.Need + " was enough"
		}
		fmt.Fprintf(out, "  %s: -$%.2f  %s  $%.2f left\n\n", verdict, r.USD, budget.Bar(c.Spec.MonthlyBudgetUSD-left, c.Spec.MonthlyBudgetUSD, 24), left)
	}
	if left > 0 {
		fmt.Fprintf(out, "Month survived with $%.2f left. Score: %.0f\n", left, left*100)
	} else {
		fmt.Fprintf(out, "Quota on fire: $%.2f over budget.\n", -left)
	}
	return left, nil
}
