package repo

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v29/github"
	"k8s.io/klog/v2"
)

const dateForm = "2006-01-02"

var (
	ignorePathRe = regexp.MustCompile(`go\.mod|go\.sum|vendor/|third_party|ignore|schemas/v\d|schema/v\d|Gopkg.lock|.DS_Store`)
	truncRe      = regexp.MustCompile(`changelog|CHANGELOG|Gopkg.toml`)
	commentRe    = regexp.MustCompile(`<!--.*?>`)
)

// FilteredFiles returns a filtered list of files modified by a PR
func FilteredFiles(ctx context.Context, c *github.Client, org string, project string, num int) ([]*github.CommitFile, error) {
	klog.Infof("Fetching file list for #%d", num)

	var files []*github.CommitFile
	changed, _, err := c.PullRequests.ListFiles(ctx, org, project, num, &github.ListOptions{})
	if err != nil {
		return files, err
	}

	for _, cf := range changed {
		if ignorePathRe.MatchString(cf.GetFilename()) {
			klog.Infof("ignoring %s", cf.GetFilename())
			continue
		}
		files = append(files, cf)
	}
	return files, err
}

// ListPulls returns a list of pull requests in a project
func ListPulls(ctx context.Context, c *github.Client, org string, project string, since time.Time, until time.Time, users []string) ([]*github.PullRequest, error) {
	var result []*github.PullRequest

	opts := &github.PullRequestListOptions{
		State:     "closed",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	matchUser := map[string]bool{}
	for _, u := range users {
		matchUser[strings.ToLower(u)] = true
	}

	klog.Infof("Gathering pull requests for %s/%s, users=%q: %+v", org, project, users, opts)
	for page := 1; page != 0; {
		opts.ListOptions.Page = page
		prs, resp, err := c.PullRequests.List(ctx, org, project, opts)
		if err != nil {
			klog.Errorf("list failed, retrying: %v", err)
			time.Sleep(time.Second * 1)
			prs, resp, err = c.PullRequests.List(ctx, org, project, opts)
			if err != nil {
				return result, err
			}
		}

		klog.Infof("Processing page %d of %s/%s pull request results ...", page, org, project)

		page = resp.NextPage
		for _, pr := range prs {
			if pr.GetClosedAt().After(until) {
				klog.Infof("PR#d closed at %s", pr.GetNumber(), pr.GetUpdatedAt())
				continue
			}

			if pr.GetUpdatedAt().Before(since) {
				klog.Infof("Hit PR#%d updated at %s", pr.GetNumber(), pr.GetUpdatedAt())
				page = 0
				break
			}

			uname := strings.ToLower(pr.GetUser().GetLogin())
			if len(matchUser) > 0 && !matchUser[uname] {
				continue
			}

			if pr.GetState() != "closed" {
				klog.Infof("Skipping PR#%d by %s (state=%q)", pr.GetNumber(), pr.GetUser().GetLogin(), pr.GetState())
				continue
			}

			klog.Infof("Fetching PR #%d by %s: %q", pr.GetNumber(), pr.GetUser().GetLogin(), pr.GetTitle())

			fullPR, _, err := c.PullRequests.Get(ctx, org, project, pr.GetNumber())
			if err != nil {
				klog.Errorf("pull failed, retrying: %v", err)
				time.Sleep(time.Second * 1)
				fullPR, _, err = c.PullRequests.Get(ctx, org, project, pr.GetNumber())
				if err != nil {
					klog.Errorf("unable to get details for %d: %v", pr.GetNumber(), err)
					// Accept partial credit
					result = append(result, pr)
					continue
				}
			}

			if !fullPR.GetMerged() || fullPR.GetMergeCommitSHA() == "" {
				klog.Infof("#%d was not merged, skipping", pr.GetNumber())
				continue
			}
			result = append(result, fullPR)
		}
	}
	klog.Infof("Returning %d pull request results", len(result))
	return result, nil
}

// PRSummary is a summary of a single PR
type PRSummary struct {
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
}

// PullSummary converts GitHub PR data into a summarized view
func PullSummary(prs map[*github.PullRequest][]*github.CommitFile, since time.Time, until time.Time, includeFileInfo bool) ([]*PRSummary, error) {
	sum := []*PRSummary{}
	seen := map[string]bool{}

	for pr, files := range prs {
		if seen[pr.GetHTMLURL()] {
			klog.Infof("skipping seen issue: %s", pr.GetHTMLURL())
			continue
		}
		seen[pr.GetHTMLURL()] = true

		_, project := ParseURL(pr.GetHTMLURL())
		body := pr.GetBody()
		body = commentRe.ReplaceAllString(body, "")

		if len(body) > 240 {
			body = body[0:240] + "..."
		}
		t := pr.GetMergedAt()
		// Often the merge timestamp is empty :(
		if t.IsZero() {
			t = pr.GetClosedAt()
		}

		if t.After(until) {
			klog.Infof("skipping %s - closed at %s, after %s", pr.GetHTMLURL(), t, until)
			continue
		}

		if t.Before(since) {
			klog.Infof("skipping %s - closed at %s, before %s", pr.GetHTMLURL(), t, since)
			continue
		}

		added := 0
		paths := []string{}
		deleted := 0

		if !includeFileInfo {
			added = pr.GetAdditions()
			deleted = pr.GetDeletions()
		} else {
			for _, f := range files {
				// These files are mostly auto-generated
				if truncRe.MatchString(f.GetFilename()) && f.GetAdditions() > 10 {
					klog.Infof("truncating %s from %d to %d lines added", f.GetFilename(), f.GetAdditions(), 10)
					added += 10
				} else {
					added += f.GetAdditions()
				}
				deleted += f.GetDeletions()
				paths = append(paths, f.GetFilename())
			}
		}

		sum = append(sum, &PRSummary{
			URL:         pr.GetHTMLURL(),
			Date:        t.Format(dateForm),
			Project:     project,
			Type:        prType(files),
			Title:       pr.GetTitle(),
			User:        pr.GetUser().GetLogin(),
			Delta:       added + deleted,
			Added:       added,
			Deleted:     deleted,
			FilesTotal:  pr.GetChangedFiles(),
			Files:       strings.Join(paths, "\n"),
			Description: body,
		})
	}
	return sum, nil
}

// prType returns what kind of PR it thinks this may be
func prType(files []*github.CommitFile) string {
	result := ""
	for _, cf := range files {
		f := cf.GetFilename()
		ext := strings.TrimLeft(filepath.Ext(f), ".")

		if strings.Contains(filepath.Dir(f), "docs/") || strings.Contains(filepath.Dir(f), "examples/") || strings.Contains(filepath.Dir(f), "site/") {
			if result == "" {
				result = "docs"
			}
			klog.Infof("%s: %s", f, result)
			continue
		}

		if strings.Contains(f, "test") || strings.Contains(f, "integration") {
			if result == "" {
				result = "tests"
			}
			klog.Infof("%s: %s", f, result)
			continue
		}

		if ext == "md" && result == "" {
			result = "docs"
		}

		if ext == "go" || ext == "java" || ext == "cpp" || ext == "py" || ext == "c" || ext == "rs" {
			result = "backend"
		}

		if ext == "ts" || ext == "js" || ext == "html" {
			result = "frontend"
		}

		klog.Infof("%s (ext=%s): %s", f, ext, result)
	}

	if result == "" {
		return "unknown"
	}
	return result
}
