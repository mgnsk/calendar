package timestamp

import (
	"fmt"
)

// FormatDay returns day with the ordinal suffix for day.
func FormatDay(day int) string {
	return fmt.Sprintf("%d%s", day, getDaySuffix(day))
}

func getDaySuffix(n int) string {
	if n >= 11 && n <= 13 {
		return "th"
	}

	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
