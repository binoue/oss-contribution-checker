package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/muesli/termenv"
)

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

// customRenderJSON encodes the JSON output and prints it.
func customRenderJSON(g []GithubIssue) error {
	// output, err := json.MarshalIndent(m, "", " ")
	// if err != nil {
	// 	return fmt.Errorf("error formatting the json output: %s", err)
	// }

	// fmt.Println(string(output))
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

		// Repo/Year base summary
		{ID: "issue_num", Name: "issue count", SortIndex: 4, Width: 3},
		{ID: "pr_num", Name: "PR count", SortIndex: 5, Width: 3},
		{ID: "issue_percent", Name: "issue%", SortIndex: 6},
		{ID: "pr_percent", Name: "PR%", SortIndex: 7},
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
		{Number: 5, Hidden: !inColumns(cols, 5)},
		{Number: 6, Hidden: !inColumns(cols, 6)},
		{Number: 7, Hidden: !inColumns(cols, 7), Transformer: barTransformer, WidthMax: int(float64(twidth) * 0.35), Align: text.AlignLeft, AlignHeader: text.AlignLeft},
		{Number: 8, Hidden: !inColumns(cols, 8), Transformer: barTransformer, WidthMax: int(float64(twidth) * 0.35), Align: text.AlignLeft, AlignHeader: text.AlignLeft},
	})

	headers := table.Row{}
	for _, v := range customColumns {
		headers = append(headers, v.Name)
	}
	tab.AppendHeader(headers)

	// count issues/pr based on repo
	repoPRMap := make(map[string]int)
	repoIssueMap := make(map[string]int)
	for _, v := range g {
		if _, ok := repoPRMap[v.project]; !ok {
			repoPRMap[v.project] = 0
		}
		if _, ok := repoIssueMap[v.project]; !ok {
			repoIssueMap[v.project] = 0
		}
		if v.isPR {
			val, _ := repoPRMap[v.project]
			repoPRMap[v.project] = val + 1
			continue
		}
		val, _ := repoIssueMap[v.project]
		repoIssueMap[v.project] = val + 1
	}

	// count issues/pr based on year
	yearPRMap := make(map[int]int)
	yearIssueMap := make(map[int]int)
	startYear := 9999
	endYear := 0
	for _, v := range g {
		year, _ := strconv.Atoi(v.year)
		if year < startYear {
			startYear = year
		}
		if endYear < year {
			endYear = year
		}
		if _, ok := yearPRMap[year]; !ok {
			yearPRMap[year] = 0
		}
		if _, ok := yearIssueMap[year]; !ok {
			yearIssueMap[year] = 0
		}

		if v.isPR {
			val, _ := yearPRMap[year]
			yearPRMap[year] = val + 1
			continue
		}
		val, _ := yearIssueMap[year]
		yearIssueMap[year] = val + 1
	}

	// filling no data year
	if startYear != 0 {
		for i := startYear; i < endYear; i++ {
			if _, ok := yearPRMap[i]; !ok {
				yearPRMap[i] = 0
			}
			if _, ok := yearIssueMap[i]; !ok {
				yearIssueMap[i] = 0
			}
		}
	}

	if params.repo && params.summary {
		totalRepoPRCount := 0
		totalRepoIssueCount := 0

		for _, v := range repoPRMap {
			totalRepoPRCount += v
		}
		for _, v := range repoIssueMap {
			totalRepoIssueCount += v
		}

		for k, v := range repoPRMap {
			tab.AppendRow([]interface{}{
				"",              // year
				"",              //title
				k,               // project name
				false,           // isPR
				repoIssueMap[k], // issue_num
				v,               // pr_num
				float64(repoIssueMap[k]) / float64(totalRepoIssueCount), // issue_percent
				float64(repoPRMap[k]) / float64(totalRepoPRCount),       // pr_percent
			})
		}
	} else if params.summary {
		totalYearPRCount := 0
		totalYearIssueCount := 0

		for _, v := range yearPRMap {
			totalYearPRCount += v
		}
		for _, v := range yearIssueMap {
			totalYearIssueCount += v
		}

		for k, v := range yearPRMap {
			tab.AppendRow([]interface{}{
				k,               // year
				"",              //title
				"",              // project name
				false,           // isPR
				yearIssueMap[k], // issue_num
				v,               // pr_num
				float64(yearIssueMap[k]) / float64(totalYearIssueCount), // issue_percent
				float64(yearPRMap[k]) / float64(totalYearPRCount),       // pr_percent
			})
		}
	} else {
		for _, v := range g {
			tab.AppendRow([]interface{}{
				termenv.String(v.year).Foreground(theme.colorBlue),
				v.title,
				v.project,
				isPR(v.isPR),
			})
		}
	}

	if tab.Length() == 0 {
		return
	}

	if params.summary && params.repo {
		tab.SetTitle("Your %d contributed projects", tab.Length())
	} else if params.summary {
		tab.SetTitle("Your yearly contribution")
	} else {
		tab.SetTitle("Your %d Issues/PRs", tab.Length())
	}

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
