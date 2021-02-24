# pullsheet

pullsheet generates a CSV (comma separated values) file containing metadata about GitHub PR's merged or reviewed by a user or group of users across a list of GitHub repositories. 

This tool was created as a brain-tickler for what PR's to discuss when asking for that big promotion.

## Usage

`go run pullsheet --repos <repository> --since 2006-01-02 --token <github token> [--users=<user>]`

You will need a GitHub authentication token from https://github.com/settings/tokens

## Example: Merged PRs for 1 person across repos

`go run pullsheet.go --repos kubernetes/minikube,GoogleContainerTools/skaffold --since 2019-10-01 --token XXX --user someone > someone.csv`

## Example: Merged PR Reviews for all users in a repo

`go run pullsheet.go --repos kubernetes/minikube --reviews --since 2020-12-24 --token XXX > reviews.csv`

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
