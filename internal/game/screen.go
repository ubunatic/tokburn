package game

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	flames "ubunatic.com/tokburn/design/flames-sol"
	final "ubunatic.com/tokburn/design/game-final"
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

// Digit strokes occupy five rows and five columns; two digits fit the side panel.
var taskDigits = [10][5]string{
	{" ███ ", "█   █", "█   █", "█   █", " ███ "},
	{"  █  ", " ██  ", "  █  ", "  █  ", " ███ "},
	{" ███ ", "█   █", "   █ ", "  █  ", "█████"},
	{"████ ", "    █", " ███ ", "    █", "████ "},
	{"█  █ ", "█  █ ", "█████", "   █ ", "   █ "},
	{"█████", "█    ", "████ ", "    █", "████ "},
	{" ███ ", "█    ", "████ ", "█   █", " ███ "},
	{"█████", "    █", "   █ ", "  █  ", "  █  "},
	{" ███ ", "█   █", " ███ ", "█   █", " ███ "},
	{" ███ ", "█   █", " ████", "    █", " ███ "},
}

func gameScreen(c budget.Calculator, card Card, index, total int, left, limit float64, jokers, combo int, revealed bool) canvas {
	s := parseCanvas(final.Game)
	// Preserve the sketch's padding attributes when replacing each field.
	field := func(y, x, width int, text string, style cell) {
		s.put(y, x, width, "", cell{fg: 252, bg: 16})
		s.put(y, x, min(width, len([]rune(text))), text, style)
	}
	for y := 0; y < 5; y++ {
		text := ""
		for _, digit := range fmt.Sprintf("%02d", index) {
			if len(text) > 0 {
				text += "  "
			}
			text += taskDigits[digit-'0'][y]
		}
		s.put(5+y, 4, 12, text, cell{fg: 220, bg: 16, bold: true})
	}
	field(11, 4, 18, fmt.Sprintf("TASK / %d", total), cell{fg: 245, bg: 16})
	field(14, 4, 18, fmt.Sprintf("$%.2f", left), cell{fg: 220, bg: 16, bold: true})
	field(17, 4, 18, fmt.Sprintf("◆ %d JOKER", jokers), cell{fg: 81, bg: 16})
	slots := max(0, min(c.Spec.JokerComboAnswers, 6))
	marks := max(0, min(combo, slots))
	meter := strings.TrimSpace(strings.Repeat("● ", marks) + strings.Repeat("○ ", slots-marks))
	field(19, 4, 18, "COMBO "+meter, cell{fg: 214, bg: 16})
	lines := []string{""}
	for _, word := range strings.Fields(card.Task) {
		i := len(lines) - 1
		if len([]rune(lines[i]))+1+len([]rune(word)) > 50 && lines[i] != "" {
			lines = append(lines, word)
		} else {
			if lines[i] != "" {
				lines[i] += " "
			}
			lines[i] += word
		}
	}
	for i := 0; i < 4; i++ {
		text := ""
		if i < len(lines) {
			text = lines[i]
		}
		field(5+i, 25, 50, text, cell{fg: 231, bg: 16, bold: true})
	}
	field(9, 25, 50, fmt.Sprintf("%dk tokens × %d runs", card.Tokens/1000, card.Runs), cell{fg: 245, bg: 16})
	for i, tier := range []string{"low", "mid", "high"} {
		e, _ := c.Estimate(card.Tokens, tier)
		text := fmt.Sprintf("$%.2f", e.USD*float64(card.Runs))
		// Centre prices in their fixed eleven-cell tiles, with the sketch's left bias.
		padding := max(0, (12-len([]rune(text)))/2)
		s.put(15, 25+13*i, 11, strings.Repeat(" ", padding)+text, cell{fg: 16, bg: []int{71, 178, 167}[i]})
	}
	if revealed {
		s.put(15, 64, 11, "   tier", cell{fg: 16, bg: 44})
		text := card.Need
		s.put(16, 64, 11, strings.Repeat(" ", (11-len(text))/2)+text, cell{fg: 16, bg: 44})
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
