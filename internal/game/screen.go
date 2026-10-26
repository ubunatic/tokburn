package game

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	flames "ubunatic.com/tokburn/design/flames-sol"
	"ubunatic.com/tokburn/design/opus"
	"ubunatic.com/tokburn/internal/budget"
)

// A canvas keeps template glyphs and SGR attributes together, so editing a
// dynamic field cannot move a border or change its neighbour's colour.
type cell struct {
	glyph  rune
	fg, bg int
	bold   bool
}
type canvas [][]cell

func parseCanvas(s string) canvas {
	var rows canvas
	row := []cell{}
	style := cell{fg: -1, bg: -1}
	for len(s) > 0 {
		if strings.HasPrefix(s, "\x1b[") {
			end := strings.IndexByte(s, 'm')
			if end < 0 {
				break
			}
			codes := strings.Split(s[2:end], ";")
			for i := 0; i < len(codes); i++ {
				n, _ := strconv.Atoi(codes[i])
				switch n {
				case 0:
					style = cell{fg: -1, bg: -1}
				case 1:
					style.bold = true
				case 22:
					style.bold = false
				case 39:
					style.fg = -1
				case 49:
					style.bg = -1
				case 38, 48:
					if i+2 < len(codes) && codes[i+1] == "5" {
						v, _ := strconv.Atoi(codes[i+2])
						if n == 38 {
							style.fg = v
						} else {
							style.bg = v
						}
						i += 2
					}
				}
			}
			s = s[end+1:]
			continue
		}
		r, n := utf8.DecodeRuneInString(s)
		s = s[n:]
		if r == '\n' {
			rows = append(rows, row)
			row = nil
			continue
		}
		c := style
		c.glyph = r
		row = append(row, c)
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	return rows
}
func (c canvas) put(y, x, width int, text string, style cell) {
	runes := []rune(text)
	for i := 0; i < width && x+i < len(c[y]); i++ {
		v := style
		v.glyph = ' '
		if i < len(runes) {
			v.glyph = runes[i]
		}
		c[y][x+i] = v
	}
}
func (c canvas) ansi() string {
	var b strings.Builder
	for y, row := range c {
		fmt.Fprintf(&b, "\x1b[%d;1H", y+1)
		prev := cell{fg: -2, bg: -2}
		for _, v := range row {
			if v.fg != prev.fg || v.bg != prev.bg || v.bold != prev.bold {
				b.WriteString("\x1b[0m")
				if v.bold {
					b.WriteString("\x1b[1m")
				}
				if v.fg >= 0 {
					fmt.Fprintf(&b, "\x1b[38;5;%dm", v.fg)
				}
				if v.bg >= 0 {
					fmt.Fprintf(&b, "\x1b[48;5;%dm", v.bg)
				}
			}
			b.WriteRune(v.glyph)
			prev = v
		}
	}
	b.WriteString("\x1b[0m")
	return b.String()
}
func ink(fg int, bold bool) cell { return cell{fg: fg, bg: -1, bold: bold} }

var opusTitle = parseCanvas(opus.Title)

var titleFrames = func() []canvas {
	names, err := flames.Frames.ReadDir(".")
	if err != nil {
		panic(err)
	}
	var frames []canvas
	for _, name := range names {
		data, err := flames.Frames.ReadFile(name.Name())
		if err != nil {
			panic(err)
		}
		frames = append(frames, parseCanvas(string(data)))
	}
	return frames
}()

func titleScreen(s budget.Spec, selected, count int) canvas {
	return titleScreenFrame(s, selected, count, 0)
}

func titleScreenFrame(s budget.Spec, selected, count, frame int) canvas {
	original := opusTitle
	c := parseCanvas(strings.Repeat(strings.Repeat(" ", 80)+"\n", 24))
	for y, row := range titleFrames[frame%len(titleFrames)] {
		if len(row) > 48 {
			panic(fmt.Sprintf("flame frame row %d is %d cells wide; maximum is 48", y+1, len(row)))
		}
		copy(c[y][16:64], row)
	}
	c[12], c[13] = original[9], original[10]
	for y := 12; y <= 18; y++ {
		c[y+2] = original[y]
	}
	c[21], c[23] = original[20], original[23]
	labels := []string{fmt.Sprintf("Start (%d tasks)", count), fmt.Sprintf("Start (%d tasks)", s.GameRoundTasks), "Exit"}
	for i, label := range labels {
		text := "   " + label
		style := ink(252, false)
		if i == 2 {
			style.fg = 245
		}
		if i == selected {
			text = " ▶ " + label
			style = cell{fg: 16, bg: 214, bold: true}
		}
		if i == 0 {
			text = fmt.Sprintf("%-29s[-] [+]", text)
		}
		c.put(16+i, 22, 36, text, style)
	}
	return c
}
func gameScreen(c budget.Calculator, card Card, index, total int, left, limit float64, jokers, combo int, revealed bool) canvas {
	s := parseCanvas(opus.Game)
	header := fmt.Sprintf("  QUOTA DEFENSE%*s", 68-len("  QUOTA DEFENSE"), fmt.Sprintf("TASK %d / %d", index, total))
	s.put(0, 11, 68, header, cell{fg: 223, bg: 52})
	// Keep the gauge, joker and combo columns fixed for every budget amount.
	amount := fmt.Sprintf("$%.2f", left)
	width := max(7, len(amount))
	s.put(1, 8, width, amount, ink(226, true))
	s.put(1, 8+width, 1, " ", ink(-1, false))
	gaugeWidth := 37 - width
	ratio := math.Max(0, math.Min(1, left/limit))
	bars := int(math.Round(float64(gaugeWidth) * ratio))
	s.put(1, 9+width, gaugeWidth, strings.Repeat("█", bars)+strings.Repeat("░", gaugeWidth-bars), ink(208, false))
	s.put(1, 46, 5, fmt.Sprintf(" %3.0f%%", ratio*100), ink(250, false))
	s.put(1, 60, 3, fmt.Sprintf("◆ %d", jokers), ink(51, true))
	comboSlots := min(c.Spec.JokerComboAnswers, 6)
	comboMarks := min(combo, comboSlots)
	s.put(1, 72, comboSlots, strings.Repeat("●", comboMarks)+strings.Repeat("○", max(0, comboSlots-comboMarks)), ink(240, false))
	// Tier labels use the spec rates; card costs below include the surcharge.
	for i, t := range c.Spec.Tiers {
		if i >= 3 {
			break
		}
		x := []int{4, 22, 40}[i]
		w := []int{15, 15, 16}[i]
		s.put(2, x, w, fmt.Sprintf(" %s $%.2f/M   ", t.Name, t.USDPerMillion), ink([]int{71, 178, 167}[i], false))
	}
	label := fmt.Sprintf(" TASK %d ", index)
	s.put(5, 4, 62, label+strings.Repeat("─", 62-len(label)), ink(208, false))
	s.put(5, 4, len(label), label, ink(231, true))
	if revealed {
		s.put(5, 66, 10, fmt.Sprintf(" tier %-4s", card.Need), ink(51, true))
	}
	words := strings.Fields(card.Task)
	lines := []string{""}
	for _, word := range words {
		i := len(lines) - 1
		if len([]rune(lines[i]))+1+len([]rune(word)) > 70 {
			lines = append(lines, word)
		} else {
			if lines[i] != "" {
				lines[i] += " "
			}
			lines[i] += word
		}
	}
	for i := 0; i < 3; i++ {
		text := ""
		if i < len(lines) {
			text = lines[i]
		}
		s.put(6+i, 4, 70, text, ink(231, true))
	}
	// The blank third line retains default attributes in the original sketch.
	if len(lines) < 3 {
		s.put(8, 4, 70, "", ink(-1, false))
	}
	s.put(9, 4, 72, "", ink(245, false))
	x := 4
	parts := []struct {
		text  string
		style cell
	}{{fmt.Sprintf("%dk", card.Tokens/1000), ink(252, true)}, {" tokens × ", ink(245, false)}, {fmt.Sprint(card.Runs), ink(252, true)}, {" runs  =  ", ink(245, false)}, {fmt.Sprintf("%.1fM", float64(card.Tokens)*float64(card.Runs)/1e6), ink(252, true)}, {" tokens", ink(245, false)}}
	for _, p := range parts {
		n := len([]rune(p.text))
		s.put(9, x, n, p.text, p.style)
		x += n
	}
	s.put(10, 4, 72, "cost if routed:  ", ink(245, false))
	x = 21
	for i, tier := range []string{"low", "mid", "high"} {
		e, _ := c.Estimate(card.Tokens, tier)
		text := fmt.Sprintf("%s $%.2f", tier[:1], e.USD*float64(card.Runs))
		s.put(10, x, len(text), text, ink([]int{71, 178, 167}[i], false))
		x += len(text) + 3
	}
	return s
}

func resultScreen(left, score, good float64, allMid bool) canvas {
	s := parseCanvas(opus.Result)
	text := fmt.Sprintf("Month survived with $%.2f left. Score: %.2f", left, score)
	color := 71
	if left <= 0 {
		text = fmt.Sprintf("Quota on fire: $%.2f over budget. Score: %.2f", max(-left, 0), score)
		color = 167
	}
	s.put(9, 4, 72, text, ink(color, true))
	s.put(11, 4, 72, fmt.Sprintf("Good score: %.0f or higher (perfect routing: 100).", good), ink(245, false))
	if !allMid {
		s.put(13, 4, 72, "", ink(-1, false))
	}
	return s
}
