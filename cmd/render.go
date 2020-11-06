package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
	title   string
	project string
	year    string
}

func customRenderJSON(g []GithubIssue) error {
	return nil
}

func customRenderTables(g []GithubIssue, columns []int, sortCol int, style table.Style) {
	customPrintTable("special", g, sortCol, columns, style)
}

type CustomColumn struct {
	ID        string
	Name      string
	SortIndex int
	Width     int
}

var (
	// "Mounted on", "Size", "Used", "Avail", "Use%", "Inodes", "Used", "Avail", "Use%", "Type", "Filesystem"
	// mountpoint, size, used, avail, usage, inodes, inodes_used, inodes_avail, inodes_usage, type, filesystem
	customColumns = []CustomColumn{
		{ID: "project", Name: "Project", SortIndex: 1, Width: 7},
		{ID: "title", Name: "Title", SortIndex: 5, Width: 7},
		{ID: "type", Name: "Type", SortIndex: 10},
		// {ID: "mountpoint", Name: "Mounted on", SortIndex: 1},
		// {ID: "size", Name: "Size", SortIndex: 12, Width: 7},
		// {ID: "used", Name: "Used", SortIndex: 13, Width: 7},
		// {ID: "avail", Name: "Avail", SortIndex: 14, Width: 7},
		// {ID: "usage", Name: "Use%", SortIndex: 15, Width: 6},
		// {ID: "inodes", Name: "Inodes", SortIndex: 16, Width: 7},
		// {ID: "inodes_used", Name: "Used", SortIndex: 17, Width: 7},
		// {ID: "inodes_avail", Name: "Avail", SortIndex: 18, Width: 7},
		// {ID: "inodes_usage", Name: "Use%", SortIndex: 19, Width: 6},
		// {ID: "type", Name: "Type", SortIndex: 10},
		// {ID: "filesystem", Name: "Filesystem", SortIndex: 11},
	}
)

func customPrintTable(title string, g []GithubIssue, sortBy int, cols []int, style table.Style) {
	tab := table.NewWriter()
	tab.SetAllowedRowLength(int(params.width))
	tab.SetOutputMirror(os.Stdout)
	tab.Style().Options.SeparateColumns = true
	tab.SetStyle(style)

	twidth := tableWidth(cols, tab.Style().Options.SeparateColumns)
	tab.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Hidden: !inColumns(cols, 1), WidthMax: int(float64(twidth) * 0.4)},
		{Number: 2, Hidden: !inColumns(cols, 2), WidthMax: int(float64(twidth) * 0.4), Align: text.AlignRight, AlignHeader: text.AlignRight},
		{Number: 3, Hidden: !inColumns(cols, 3), WidthMax: int(float64(twidth) * 0.4), Align: text.AlignRight, AlignHeader: text.AlignRight},
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
			// termenv.String(v.Mountpoint).Foreground(theme.colorBlue), // mounted on
			v.year,
			v.title,
			v.project,
		})
	}

	if tab.Length() == 0 {
		return
	}

	//tab.AppendFooter(table.Row{fmt.Sprintf("%d %s", tab.Length(), title)})
	sortMode := table.Asc
	if sortBy >= 12 {
		sortMode = table.AscNumeric
	}

	tab.SortBy([]table.SortBy{{Number: sortBy, Mode: sortMode}})
	tab.Render()

	return
}

// // printTable prints an individual table of mounts.
// func printTable(title string, m []Mount, sortBy int, cols []int, style table.Style) {

// 	for _, v := range m {
// 		// spew.Dump(v)

// 		var usage, inodeUsage float64
// 		if v.Total > 0 {
// 			usage = float64(v.Used) / float64(v.Total)
// 			if usage > 1.0 {
// 				usage = 1.0
// 			}
// 		}
// 		if v.Inodes > 0 {
// 			inodeUsage = float64(v.InodesUsed) / float64(v.Inodes)
// 			if inodeUsage > 1.0 {
// 				inodeUsage = 1.0
// 			}
// 		}

// 		tab.AppendRow([]interface{}{
// 			termenv.String(v.Mountpoint).Foreground(theme.colorBlue), // mounted on
// 			v.Total,      // size
// 			v.Used,       // used
// 			v.Free,       // avail
// 			usage,        // use%
// 			v.Inodes,     // inodes
// 			v.InodesUsed, // inodes used
// 			v.InodesFree, // inodes avail
// 			inodeUsage,   // inodes use%
// 			termenv.String(v.Fstype).Foreground(theme.colorGray), // type
// 			termenv.String(v.Device).Foreground(theme.colorGray), // filesystem
// 			v.Total,      // size sorting helper
// 			v.Used,       // used sorting helper
// 			v.Free,       // avail sorting helper
// 			usage,        // use% sorting helper
// 			v.Inodes,     // inodes sorting helper
// 			v.InodesUsed, // inodes used sorting helper
// 			v.InodesFree, // inodes avail sorting helper
// 			inodeUsage,   // inodes use% sorting helper
// 		})
// 	}

// 	if tab.Length() == 0 {
// 		return
// 	}

// 	suffix := "device"
// 	if tab.Length() > 1 {
// 		suffix = "devices"
// 	}
// 	tab.SetTitle("%d %s %s", tab.Length(), title, suffix)

// 	//tab.AppendFooter(table.Row{fmt.Sprintf("%d %s", tab.Length(), title)})
// 	sortMode := table.Asc
// 	if sortBy >= 12 {
// 		sortMode = table.AscNumeric
// 	}

// 	tab.SortBy([]table.SortBy{{Number: sortBy, Mode: sortMode}})
// 	tab.Render()
// }
