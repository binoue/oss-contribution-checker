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
	term  = termenv.EnvColorProfile()
	theme Theme
)

var params struct {
	summary bool
	token   string
	account string

	repo bool

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

		return showTable(results)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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

func retrieveContributionData() ([]GithubIssue, error) {
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

	var githubIssues []GithubIssue
	for _, sr := range results {
		for _, i := range sr.Issues {
			year := strconv.Itoa((i.CreatedAt).Year())
			s := strings.Split(*i.RepositoryURL, "/")
			var closed bool
			if i.ClosedAt != nil {
				closed = true
			}
			githubIssues = append(githubIssues, GithubIssue{
				title:    *i.Title,
				year:     year,
				project:  strings.Join(s[len(s)-2:], "/"),
				isPR:     i.IsPullRequest(),
				isClosed: closed,
			})
			if needToExclude(i) {
				continue
			}
		}
	}

	return githubIssues, nil
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

	rootCmd.Flags().BoolVar(&params.repo, "repo", false, "summary grouped by repo name")

	// Took from duf

	rootCmd.Flags().StringVar(&params.theme, "theme", defaultThemeName(), "color themes: dark, light")
	rootCmd.Flags().StringVar(&params.style, "style", defaultStyleName(), "style: unicode, ascii")
	rootCmd.Flags().StringVar(&params.output, "output", "", "output fields: "+strings.Join(columnIDs(), ", "))
	rootCmd.Flags().StringVar(&params.sort, "sort", "mountpoint", "sort output by: "+strings.Join(columnIDs(), ", "))
	rootCmd.Flags().UintVar(&params.width, "width", 0, "max output width")
	rootCmd.Flags().BoolVar(&params.warn, "warnings", false, "output all warnings to STDERR")
	rootCmd.Flags().BoolVar(&params.json, "json", false, "output all devices in JSON format")
}

func showTable(githubIssues []GithubIssue) error {
	var err error
	theme, err = loadTheme(params.theme)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	style, err := parseStyle(params.style)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	columns, err := parseColumns(params.output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	if len(columns) == 0 {
		if params.summary {
			if params.repo {
				columns = []int{3, 5, 6, 7, 8}
			} else {
				columns = []int{1, 2, 3, 4}
			}
		} else {
			columns = []int{1, 2, 3, 4}
		}
	}

	sortCol, err := stringToSortIndex(params.sort)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	// detect terminal width
	isTerminal := terminal.IsTerminal(int(os.Stdout.Fd()))
	if isTerminal && params.width == 0 {
		w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return err
		}
		params.width = uint(w)
	}
	if params.width == 0 {
		params.width = 80
	}

	customRenderTables(githubIssues, columns, sortCol, style)
	return nil
}
