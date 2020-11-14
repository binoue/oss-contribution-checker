package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/muesli/termenv"
)

// renderTables renders all tables.
func renderTables(m []Mount, columns []int, sortCol int, style table.Style) {
	// var local, network, fuse, special []Mount
	// hideFsMap := parseHideFs(*hideFs)

	// // sort/filter devices
	// for _, v := range m {
	// 	// skip hideFs
	// 	if _, ok := hideFsMap[v.Fstype]; ok {
	// 		continue
	// 	}
	// 	// skip autofs
	// 	if v.Fstype == "autofs" {
	// 		continue
	// 	}
	// 	// skip bind-mounts
	// 	if *hideBinds && !*all && strings.Contains(v.Opts, "bind") {
	// 		continue
	// 	}
	// 	// skip loop devices
	// 	if *hideLoops && !*all && strings.HasPrefix(v.Device, "/dev/loop") {
	// 		continue
	// 	}
	// 	// skip special devices
	// 	if v.Blocks == 0 && !*all {
	// 		continue
	// 	}
	// 	// skip zero size devices
	// 	if v.BlockSize == 0 && !*all {
	// 		continue
	// 	}

	// 	if isNetworkFs(v) {
	// 		network = append(network, v)
	// 		continue
	// 	}
	// 	if isFuseFs(v) {
	// 		fuse = append(fuse, v)
	// 		continue
	// 	}
	// 	if isSpecialFs(v) {
	// 		special = append(special, v)
	// 		continue
	// 	}

	// 	local = append(local, v)
	// }

	// // print tables
	// if !*hideLocal || *all {
	// 	printTable("local", local, sortCol, columns, style)
	// }
	// if !*hideNetwork || *all {
	// 	printTable("network", network, sortCol, columns, style)
	// }
	// if !*hideFuse || *all {
	// 	printTable("FUSE", fuse, sortCol, columns, style)
	// }
	// if !*hideSpecial || *all {
	// 	printTable("special", special, sortCol, columns, style)
	// }
	var special []Mount
	printTable("special", special, sortCol, columns, style)
}

// renderJSON encodes the JSON output and prints it.
func renderJSON(m []Mount) error {
	output, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return fmt.Errorf("error formatting the json output: %s", err)
	}

	fmt.Println(string(output))
	return nil
}

// parseColumns parses the supplied output flag into a slice of column indices.
func parseColumns(cols string) ([]int, error) {
	var i []int

	s := strings.Split(cols, ",")
	for _, v := range s {
		v = strings.TrimSpace(v)
		if len(v) == 0 {
			continue
		}

		col, err := stringToColumn(v)
		if err != nil {
			return nil, err
		}

		i = append(i, col)
	}

	return i, nil
}

// parseStyle converts user-provided style option into a table.Style.
func parseStyle(styleOpt string) (table.Style, error) {
	switch styleOpt {
	case "unicode":
		return table.StyleRounded, nil
	case "ascii":
		return table.StyleDefault, nil
	default:
		return table.Style{}, fmt.Errorf("Unknown style option: %s", styleOpt)
	}
}

// parseHideFs parses the supplied hide-fs flag into a map of fs types which should be skipped.
func parseHideFs(hideFs string) map[string]struct{} {
	hideMap := make(map[string]struct{})
	for _, fs := range strings.Split(hideFs, ",") {
		fs = strings.TrimSpace(fs)
		if len(fs) == 0 {
			continue
		}
		hideMap[fs] = struct{}{}
	}
	return hideMap
}

type GithubIssue struct {
	title    string
	project  string
	year     string
	isPR     bool
	isClosed bool
	// isMerged bool
}

func customRenderJSON(g []GithubIssue) error {
	return nil
}

func customRenderTables(g []GithubIssue, columns []int, sortCol int, style table.Style) {
	customPrintTable(g, sortCol, columns, style)
}

type CustomColumn struct {
	ID        string
	Name      string
	SortIndex int
	Width     int
}

var (
	customColumns = []CustomColumn{
		{ID: "year", Name: "Year", SortIndex: 1, Width: 7},
		{ID: "title", Name: "Title", SortIndex: 4},
		{ID: "repo", Name: "Repo", SortIndex: 3},
		{ID: "pr", Name: "PR", SortIndex: 2, Width: 3},
	}
)

func customPrintTable(g []GithubIssue, sortBy int, cols []int, style table.Style) {
	tab := table.NewWriter()
	tab.SetAllowedRowLength(int(params.width))
	tab.SetOutputMirror(os.Stdout)
	tab.Style().Options.SeparateColumns = true
	tab.SetStyle(style)

	twidth := customTableWidth(cols, tab.Style().Options.SeparateColumns, customColumns)
	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Hidden: !inColumns(cols, 1)},
		{Number: 2, Hidden: !inColumns(cols, 2), WidthMax: int(float64(twidth) * 0.7), Align: text.AlignLeft, AlignHeader: text.AlignLeft},
		{Number: 3, Hidden: !inColumns(cols, 3), WidthMax: int(float64(twidth) * 0.3), Align: text.AlignLeft, AlignHeader: text.AlignLeft},
		{Number: 4, Hidden: !inColumns(cols, 4)},

		// {Number: 2, Hidden: !inColumns(cols, 2), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		// {Number: 3, Hidden: !inColumns(cols, 3), Transformer: sizeTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		// {Number: 4, Hidden: !inColumns(cols, 4), Transformer: spaceTransformer, Align: text.AlignRight, AlignHeader: text.AlignRight},
		// {Number: 5, Hidden: !inColumns(cols, 5), Transformer: barTransformer, AlignHeader: text.AlignCenter},
		// {Number: 6, Hidden: !inColumns(cols, 6), Align: text.AlignRight, AlignHeader: text.AlignRight},
		// {Number: 7, Hidden: !inColumns(cols, 7), Align: text.AlignRight, AlignHeader: text.AlignRight},
		// {Number: 8, Hidden: !inColumns(cols, 8), Align: text.AlignRight, AlignHeader: text.AlignRight},
		// {Number: 9, Hidden: !inColumns(cols, 9), Transformer: barTransformer, AlignHeader: text.AlignCenter},
		// {Number: 10, Hidden: !inColumns(cols, 10), WidthMax: int(float64(twidth) * 0.2)},
		// {Number: 11, Hidden: !inColumns(cols, 11), WidthMax: int(float64(twidth) * 0.4)},
		// {Number: 12, Hidden: true}, // sortBy helper for size
		// {Number: 13, Hidden: true}, // sortBy helper for used
		// {Number: 14, Hidden: true}, // sortBy helper for avail
		// {Number: 15, Hidden: true}, // sortBy helper for usage
		// {Number: 16, Hidden: true}, // sortBy helper for inodes size
		// {Number: 17, Hidden: true}, // sortBy helper for inodes used
		// {Number: 18, Hidden: true}, // sortBy helper for inodes avail
		// {Number: 19, Hidden: true}, // sortBy helper for inodes usage
	})

	headers := table.Row{}
	for _, v := range customColumns {
		headers = append(headers, v.Name)
	}
	tab.AppendHeader(headers)

	for _, v := range g {
		tab.AppendRow([]interface{}{
			termenv.String(v.year).Foreground(theme.colorBlue),
			v.title,
			v.project,
			isPR(v.isPR),
		})
	}

	if tab.Length() == 0 {
		return
	}

	tab.SetTitle("Your %d Issues/PRs", tab.Length())

	sortMode := table.Asc
	if sortBy >= 12 {
		sortMode = table.AscNumeric
	}

	tab.SortBy([]table.SortBy{{Number: sortBy, Mode: sortMode}})
	tab.Render()

	return
}

func isPR(b bool) string {
	if b {
		return "○"
	}
	return "-"
}
