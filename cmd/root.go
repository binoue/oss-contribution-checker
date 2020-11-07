package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/muesli/termenv"
	"github.com/russross/blackfriday"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

var (
	Version   = ""
	CommitSHA = ""

	term  = termenv.EnvColorProfile()
	theme Theme

	// all         = flag.Bool("all", false, "include pseudo, duplicate, inaccessible file systems")
	// hideLocal   = flag.Bool("hide-local", false, "hide local devices")
	// hideNetwork = flag.Bool("hide-network", false, "hide network devices")
	// hideFuse    = flag.Bool("hide-fuse", false, "hide fuse devices")
	// hideSpecial = flag.Bool("hide-special", false, "hide special devices")
	// hideLoops   = flag.Bool("hide-loops", true, "hide loop devices")
	// hideBinds   = flag.Bool("hide-binds", true, "hide bind mounts")
	// hideFs      = flag.String("hide-fs", "", "hide specific filesystems, separated with commas")

	// output   = flag.String("output", "", "output fields: "+strings.Join(columnIDs(), ", "))
	// sortBy   = flag.String("sort", "mountpoint", "sort output by: "+strings.Join(columnIDs(), ", "))
	// width    = flag.Uint("width", 0, "max output width")
	// themeOpt = flag.String("theme", defaultThemeName(), "color themes: dark, light")
	// styleOpt = flag.String("style", defaultStyleName(), "style: unicode, ascii")

	// inodes     = flag.Bool("inodes", false, "list inode information instead of block usage")
	// jsonOutput = flag.Bool("json", false, "output all devices in JSON format")
	// warns      = flag.Bool("warnings", false, "output all warnings to STDERR")
	// version    = flag.Bool("version", false, "display version")
)
var params struct {
	summary bool
	token   string
	account string

	theme  string
	style  string
	output string
	sort   string
	width  uint
	warn   bool
	json   bool
}

type Token struct {
	GithubToken string `yaml:"github_token"`
}

var excludeOrgs = []string{"cybozu"}
var years = []string{"2017", "2018", "2019", "2020"}

var rootCmd = &cobra.Command{
	Use:   "oss-contribution-checker",
	Short: "oss-contribution-checker",
	Long:  `"oss-contribution-checker is a tool for showing your OSS contributions.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if params.account == "" {
			return errors.New("account name is not specified")
		}
		err := setToken()
		if err != nil {
			return err
		}
		results, err := retrieveContributionData()
		if err != nil {
			return err
		}
		if params.summary {
			err = showSummery(results)
			if err != nil {
				return err
			}
			return nil
		}
		showTable(results)
		return nil
	},
}

func showSummery(searchResults []*github.IssuesSearchResult) error {
	issueCountMap := make(map[string]int)
	for _, y := range years {
		issueCountMap[y] = 0
	}
	prCountMap := make(map[string]int)
	for _, y := range years {
		prCountMap[y] = 0
	}
	excludeOrgs = append(excludeOrgs, params.account)

	for _, sr := range searchResults {
		for _, i := range sr.Issues {
			year := strconv.Itoa((i.CreatedAt).Year())
			input := fmt.Sprintf("title: %v, year: %v, repositoryURL: %v, needToExclude: %v\n", *i.Title, year, *i.RepositoryURL, needToExclude(i))
			output := blackfriday.Run([]byte(input), blackfriday.WithExtensions(blackfriday.Tables))
			fmt.Println(string(output))
			if needToExclude(i) {
				continue
			}
			if i.IsPullRequest() {
				prCountMap[year] += 1
				continue
			}
			issueCountMap[year] += 1
		}
	}

	fmt.Printf("\nSummery:\n")
	fmt.Println("# of Issues:")
	for _, y := range years {
		fmt.Printf("%v,%v\n", y, issueCountMap[y])
	}
	fmt.Println("# of PRs:")
	for _, y := range years {
		fmt.Printf("%v,%v\n", y, prCountMap[y])
	}
	return nil
}

func needToExclude(issue *github.Issue) bool {
	for _, o := range excludeOrgs {
		if strings.Contains(*issue.RepositoryURL, o) {
			return true
		}
	}
	return false
}

func retrieveContributionData() ([]*github.IssuesSearchResult, error) {
	ctx := context.Background()
	token := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: params.token},
	)
	c := oauth2.NewClient(ctx, token)
	gc := github.NewClient(c)

	query := "author:" + params.account
	opts := github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	// pagenation
	var results []*github.IssuesSearchResult
	for {
		result, resp, err := gc.Search.Issues(ctx, query, &opts)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		time.Sleep(time.Duration(1))
	}
	return results, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setToken() error {
	if params.token != "" {
		return nil
	}
	b, err := ioutil.ReadFile("token.txt")
	if err == nil {
		params.token = strings.TrimSuffix(string(b), "\n")
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	b, err = ioutil.ReadFile(home + "/.git-neco.yml")
	if err != nil {
		return err
	}
	var t Token
	err = yaml.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	params.token = strings.TrimSuffix(string(t.GithubToken), "\n")
	if params.token == "" {
		return errors.New("failed to get token")
	}
	return nil
}

func init() {
	rootCmd.Flags().BoolVar(&params.summary, "summary", false, "show summary")
	rootCmd.Flags().StringVar(&params.token, "token", "", "github token")
	rootCmd.Flags().StringVar(&params.account, "account", "", "your github account name")

	rootCmd.Flags().StringVar(&params.theme, "theme", defaultThemeName(), "color themes: dark, light")
	rootCmd.Flags().StringVar(&params.style, "style", defaultStyleName(), "style: unicode, ascii")
	rootCmd.Flags().StringVar(&params.output, "output", "", "output fields: "+strings.Join(columnIDs(), ", "))
	rootCmd.Flags().StringVar(&params.sort, "sort", "mountpoint", "sort output by: "+strings.Join(columnIDs(), ", "))
	rootCmd.Flags().UintVar(&params.width, "width", 0, "max output width")
	rootCmd.Flags().BoolVar(&params.warn, "warnings", false, "output all warnings to STDERR")
	rootCmd.Flags().BoolVar(&params.json, "json", false, "output all devices in JSON format")
}

// func main() {
// 	flag.Parse()

// 	if *version {
// 		if len(CommitSHA) > 7 {
// 			CommitSHA = CommitSHA[:7]
// 		}
// 		if Version == "" {
// 			Version = "(built from source)"
// 		}

// 		fmt.Printf("duf %s", Version)
// 		if len(CommitSHA) > 0 {
// 			fmt.Printf(" (%s)", CommitSHA)
// 		}

// 		fmt.Println()
// 		os.Exit(0)
// 	}

// 	// validate flags
func showTable(searchResults []*github.IssuesSearchResult) {
	var githubIssues []GithubIssue
	for _, sr := range searchResults {
		for _, i := range sr.Issues {
			year := strconv.Itoa((i.CreatedAt).Year())
			githubIssues = append(githubIssues, GithubIssue{
				title:   *i.Title,
				year:    year,
				project: *i.RepositoryURL,
			})
			if needToExclude(i) {
				continue
			}
		}
	}

	var err error
	theme, err = loadTheme(params.theme)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	style, err := parseStyle(params.style)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	columns, err := parseColumns(params.output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(columns) == 0 {
		columns = []int{1, 2, 3, 4, 5, 10, 11}
		// no columns supplied, use defaults
		// if *inodes {
		// 	columns = []int{1, 6, 7, 8, 9, 10, 11}
		// } else {
		// 	columns = []int{1, 2, 3, 4, 5, 10, 11}
		// }
	}

	sortCol, err := stringToSortIndex(params.sort)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// detect terminal width
	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	if isTerminal && params.width == 0 {
		w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err == nil {
			params.width = uint(w)
		}
	}
	if params.width == 0 {
		params.width = 80
	}

	// 	// read mount table
	// 	m, warnings, err := mounts()
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err)
	// 		os.Exit(1)
	// 	}

	// print out warnings
	// if *warns {
	// 	for _, warning := range warnings {
	// 		fmt.Fprintln(os.Stderr, warning)
	// 	}
	// }

	// var dummyOutput []Mount

	// // print JSON
	// if params.json {
	// 	err := renderJSON(dummyOutput)
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err)
	// 	}
	// 	return
	// }

	// print tables
	// renderTables(dummyOutput, columns, sortCol, style)
	customRenderTables(githubIssues, columns, sortCol, style)
}

// ---
// <p>title: Use firebase and mysql as its background db, year: 2020, repositoryURL: https://api.github.com/repos/banban9999/SimpleWebserver, needToExclude: true</p>

// <p>title: Use pwa, year: 2020, repositoryURL: https://api.github.com/repos/banban9999/SimpleWebserver, needToExclude: true</p>

// <p>title: test, year: 2020, repositoryURL: https://api.github.com/repos/banban9999/browser-extension, needToExclude: true</p>

// <p>title: Add card pages, year: 2020, repositoryURL: https://api.github.com/repos/banban9999/SimpleWebserver, needToExclude: true</p>

// <p>title: Fix issue #4, year: 2020, repositoryURL: https://api.github.com/repos/aNickzz/DashBot, needToExclude: false</p>

// <p>title: Greeting message is hard coded and can not use real bot name, year: 2020, repositoryURL: https://api.github.com/repos/aNickzz/DashBot, needToExclude: false</p>

// <p>title: Add reviewed comments from Neco members, year: 2020, repositoryURL: https://api.github.com/repos/banban9999/kintone-blog, needToExclude: true</p>

// <p>title: add gitignore file, year: 2019, repositoryURL: https://api.github.com/repos/banban9999/hackday, needToExclude: true</p>

// <p>title: add marble, year: 2019, repositoryURL: https://api.github.com/repos/banban9999/hackday, needToExclude: true</p>

// <p>title: Update starwarstext, year: 2019, repositoryURL: https://api.github.com/repos/banban9999/hackday, needToExclude: true</p>

// <p>title: StarwarsTextの内容改善, year: 2019, repositoryURL: https://api.github.com/repos/banban9999/hackday, needToExclude: true</p>

// <p>title: StarwarsTextの速度改善, year: 2019, repositoryURL: https://api.github.com/repos/banban9999/hackday, needToExclude: true</p>

// <p>title: StarwarsTextの背景画像変更, year: 2019, repositoryURL: https://api.github.com/repos/banban9999/hackday, needToExclude: true</p>

// <p>title: Add basic starwars text sample, year: 2019, repositoryURL: https://api.github.com/repos/banban9999/hackday, needToExclude: true</p>

// <p>title: add beginner-banban9999.yml, year: 2017, repositoryURL: https://api.github.com/repos/oss-gate/workshop, needToExclude: false</p>

// <p>title: OSS Gate Workshop: Tokyo: 2017-07-29: banban9999: vim-plugin-taskwarrior: Work log, year: 2017, repositoryURL: https://api.github.com/repos/oss-gate/workshop, needToExclude: false</p>

// Summery:
// # of Issues:
// 2017,1
// 2018,0
// 2019,0
// 2020,1
// # of PRs:
// 2017,1
// 2018,0
// 2019,0
// 2020,1
