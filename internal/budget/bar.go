package budget

import "strings"

// Bar renders value/max as a fixed-width ANSI bar: green, yellow past 70%, red past 100%.
func Bar(value, max float64, width int) string {
	if max <= 0 || width <= 0 {
		return ""
	}
	frac := value / max
	color := "\033[32m"
	switch {
	case frac >= 1:
		color = "\033[31m"
	case frac >= 0.7:
		color = "\033[33m"
	}
	filled := int(frac*float64(width) + 0.5)
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return color + strings.Repeat("█", filled) + "\033[0m" + strings.Repeat("░", width-filled)
}
