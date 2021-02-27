# pullsheet

pullsheet generates a CSV (comma separated values) & HTML output about GitHub activity across a series of repositories.

It currently supports CSV exports for:

* Merged Pull Requests: `pullsheet prs [FLAGS]`
* Pull Request Reviews: `pullsheet reviews [FLAGS]`
* Opening/Closing Issues: `pullsheet issues [FLAGS]`
* Issue Comments: `pullsheet issue-comments [FLAGS]`

As well as a new HTML leaderboard mode: `pullsheet leaderboard [FLAGS]`

This tool was created as a brain-tickler for what PR's to discuss when asking for that big promotion.

## Usage

`go run pullsheet [subcommand] --repos <repository> --since 2006-01-02 --token-path <github token path> [--users=<user>]`

You will need a GitHub authentication token from https://github.com/settings/tokens

## Example: Merged PRs for 1 person across repos

`go run pullsheet.go prs --repos kubernetes/minikube,GoogleContainerTools/skaffold --since 2019-10-01 --token-path /path/to/github/token/file --user someone > someone.csv`

## Example: Merged PR Reviews for all users in a repo

`go run pullsheet.go reviews --repos kubernetes/minikube --kind=reviews --since 2020-12-24 --token-path /path/to/github/token/file > reviews.csv`

## CSV fields

### Merged Pull Requests

```
	URL         string
	Date        string
	User        string
	Project     string
	Type        string
	Title       string
	Delta       int
	Added       int
	Deleted     int
	FilesTotal  int
	Files       string // newline delimited
	Description string
```

### Merged Pull Request Reviews

```
	URL            string
	Date           string
	Reviewer       string
	PRAuthor       string
	Project        string
	Title          string
	PRComments     int
	ReviewComments int
	Words          int
```

### Closed/Opened Issues

```
	URL     string
	Date    string
	Author  string
	Closer  string
	Project string
	Type    string
	Title   string
```

### Issue Comments

```
	URL         string
	Date        string
	Project     string
	Commenter   string
	IssueAuthor string
	IssueState  string
	Comments    int
	Words       int
	Title       string
```
