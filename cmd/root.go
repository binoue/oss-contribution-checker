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
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

var params struct {
	summary bool
	token   string
	account string
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
		err = showSummery(results)
		if err != nil {
			return err
		}
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
			fmt.Printf("title: %v, year: %v, repositoryURL: %v, needToExclude: %v\n", *i.Title, year, *i.RepositoryURL, needToExclude(i))
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
}
