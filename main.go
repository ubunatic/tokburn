// Command tokburn checks prompt sizes against the long-context price cliff
// and plays Quota Defense, a model-tier routing game.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"ubunatic.com/tokburn/internal/budget"
	"ubunatic.com/tokburn/internal/game"
	"ubunatic.com/tokburn/spec"
)

const usage = `tokburn checks prompt sizes against the long-context price cliff.

Usage:
  tokburn check [paths...]   estimate tokens and cost per tier (default: .)
  tokburn play [-seed N]     play Quota Defense
  tokburn version            print the version
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "tokburn:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	s, err := budget.LoadSpec(spec.Pricing)
	if err != nil {
		return err
	}
	c := budget.Calculator{Spec: s}
	cmd := "check"
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}
	switch cmd {
	case "check":
		return check(c, args)
	case "play":
		fs := flag.NewFlagSet("play", flag.ContinueOnError)
		seed := fs.Int64("seed", time.Now().UnixNano(), "card shuffle seed")
		if err := fs.Parse(args); err != nil {
			return err
		}
		_, err := game.Play(c, rand.New(rand.NewSource(*seed)), os.Stdin, os.Stdout)
		return err
	case "version", "--version":
		fmt.Println("tokburn", Version)
		return nil
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	}
	return fmt.Errorf("unknown command %q\n\n%s", cmd, usage)
}

func check(c budget.Calculator, paths []string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	files, err := budget.CountPaths(paths)
	if err != nil {
		return err
	}
	total := 0
	for _, f := range files {
		total += f.Tokens
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Tokens > files[j].Tokens })
	cliff := c.Spec.CliffTokens
	fmt.Printf("%d files, ~%d tokens  %s  cliff at %dk\n", len(files), total, budget.Bar(float64(total), float64(cliff), 30), cliff/1000)
	for _, f := range files[:min(5, len(files))] {
		fmt.Printf("  %8d  %s\n", f.Tokens, f.Path)
	}
	fmt.Println()
	for _, t := range c.Spec.Tiers {
		e, err := c.Estimate(total, t.Name)
		if err != nil {
			return err
		}
		mark := ""
		if e.Surcharged {
			mark = fmt.Sprintf("  x%.0f long-context surcharge", c.Spec.SurchargeMultiplier)
		}
		fmt.Printf("  %-5s $%8.4f per prompt%s\n", t.Name, e.USD, mark)
	}
	return nil
}
