package cmd

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
)

type Column struct {
	ID        string
	Name      string
	SortIndex int
	Width     int
}

// sizeTransformer makes a size human-readable.
func sizeTransformer(val interface{}) string {
	return sizeToString(val.(uint64))
}

// spaceTransformer makes a size human-readable and applies a color coding.
func spaceTransformer(val interface{}) string {
	free := val.(uint64)

	var s = termenv.String(sizeToString(free))
	switch {
	case free < 1<<30:
		s = s.Foreground(theme.colorRed)
	case free < 10*1<<30:
		s = s.Foreground(theme.colorYellow)
	default:
		s = s.Foreground(theme.colorGreen)
	}

	return s.String()
}

// barTransformer transforms a percentage into a progress-bar.
func barTransformer(val interface{}) string {
	usage := val.(float64)
	s := termenv.String()
	if usage >= 0 {
		if barWidth() > 0 {
			bw := barWidth() - 2
			s = termenv.String(fmt.Sprintf("[%s%s] %5.1f%%",
				strings.Repeat("#", int(usage*float64(bw))),
				strings.Repeat(".", bw-int(usage*float64(bw))),
				usage*100,
			))
		} else {
			s = termenv.String(fmt.Sprintf("%5.1f%%", usage*100))
		}
	}

	// apply color to progress-bar
	switch {
	case usage >= 0.9:
		s = s.Foreground(theme.colorRed)
	case usage >= 0.5:
		s = s.Foreground(theme.colorYellow)
	default:
		s = s.Foreground(theme.colorGreen)
	}
	return s.String()
}

// inColumns return true if the column with index i is in the slice of visible
// columns cols.
func inColumns(cols []int, i int) bool {
	for _, v := range cols {
		if v == i {
			return true
		}
	}

	return false
}

// barWidth returns the width of progress-bars for the given render width.
func barWidth() int {
	switch {
	case params.width < 100:
		return 0
	case params.width < 120:
		return 12
	default:
		return 22
	}
}

// tableWidth returns the required minimum table width for the given columns.
func tableWidth(cols []int, separators bool) int {
	var sw int
	if separators {
		sw = 1
	}

	twidth := int(params.width)
	for i := 0; i < len(customColumns); i++ {
		if inColumns(cols, i+1) {
			twidth -= 2 + sw + customColumns[i].Width
		}
	}

	return twidth
}

func customTableWidth(cols []int, separators bool, columns []CustomColumn) int {
	var sw int
	if separators {
		sw = 1
	}

	twidth := int(params.width)
	for i := 0; i < len(columns); i++ {
		if inColumns(cols, i+1) {
			twidth -= 2 + sw + columns[i].Width
		}
	}

	return twidth
}

// sizeToString prettifies sizes.
func sizeToString(size uint64) (str string) {
	b := float64(size)

	switch {
	case size >= 1<<60:
		str = fmt.Sprintf("%.1fE", b/(1<<60))
	case size >= 1<<50:
		str = fmt.Sprintf("%.1fP", b/(1<<50))
	case size >= 1<<40:
		str = fmt.Sprintf("%.1fT", b/(1<<40))
	case size >= 1<<30:
		str = fmt.Sprintf("%.1fG", b/(1<<30))
	case size >= 1<<20:
		str = fmt.Sprintf("%.1fM", b/(1<<20))
	case size >= 1<<10:
		str = fmt.Sprintf("%.1fK", b/(1<<10))
	default:
		str = fmt.Sprintf("%dB", size)
	}

	return
}

// stringToColumn converts a column name to its index.
func stringToColumn(s string) (int, error) {
	s = strings.ToLower(s)

	for i, v := range customColumns {
		if v.ID == s {
			return i + 1, nil
		}
	}

	return 0, fmt.Errorf("unknown column: %s (valid: %s)", s, strings.Join(columnIDs(), ", "))
}

// stringToSortIndex converts a column name to its sort index.
func stringToSortIndex(s string) (int, error) {
	s = strings.ToLower(s)

	for _, v := range customColumns {
		if v.ID == s {
			return v.SortIndex, nil
		}
	}

	return 0, fmt.Errorf("unknown column: %s (valid: %s)", s, strings.Join(columnIDs(), ", "))
}

// columnsIDs returns a slice of all column IDs.
func columnIDs() []string {
	s := make([]string, len(customColumns))
	for i, v := range customColumns {
		s[i] = v.ID
	}

	return s
}
