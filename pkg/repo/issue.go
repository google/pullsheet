package repo

import (
	"context"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/google/pullsheet/pkg/ghcache"
	"github.com/peterbourgon/diskv"
	"k8s.io/klog/v2"
)

// IssueSummary is a summary of a single PR
type IssueSummary struct {
	URL     string
	Date    string
	Author  string
	Closer  string
	Project string
	Type    string
	Title   string
}

// ClosedIssues returns a list of closed issues within a project
func ClosedIssues(ctx context.Context, dv *diskv.Diskv, c *github.Client, org string, project string, since time.Time, until time.Time, users []string) ([]*IssueSummary, error) {
	var result []*IssueSummary
	closed, err := issues(ctx, dv, c, org, project, since, until, users, "closed")
	if err != nil {
		return nil, err
	}

	for _, i := range closed {
		result = append(result, &IssueSummary{
			URL:     i.GetHTMLURL(),
			Date:    i.GetClosedAt().Format(dateForm),
			Author:  i.GetUser().GetLogin(),
			Closer:  i.GetClosedBy().GetLogin(),
			Project: project,
			Title:   i.GetTitle(),
		})
	}
	return result, nil
}

// issues returns a list of issues in a project
func issues(ctx context.Context, dv *diskv.Diskv, c *github.Client, org string, project string, since time.Time, until time.Time, users []string, state string) ([]*github.Issue, error) {
	result := []*github.Issue{}
	opts := &github.IssueListByRepoOptions{
		State:     state,
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

	klog.Infof("Gathering issues for %s/%s, users=%q: %+v", org, project, users, opts)
	for page := 1; page != 0; {
		opts.ListOptions.Page = page
		issues, resp, err := c.Issues.ListByRepo(ctx, org, project, opts)
		if err != nil {
			return result, err
		}

		klog.Infof("Processing page %d of %s/%s issue results ...", page, org, project)

		page = resp.NextPage
		for _, i := range issues {
			if i.IsPullRequest() {
				continue
			}
			if i.GetClosedAt().After(until) {
				klog.Infof("issue #d closed at %s", i.GetNumber(), i.GetUpdatedAt())
				continue
			}

			if i.GetUpdatedAt().Before(since) {
				klog.Infof("Hit issue #%d updated at %s", i.GetNumber(), i.GetUpdatedAt())
				page = 0
				break
			}

			if state != "" && i.GetState() != state {
				klog.Infof("Skipping issue #%d (state=%q)", i.GetNumber(), i.GetState())
				continue
			}

			t := issueDate(i)

			klog.Infof("Fetching #%d (closed %s, updated %s): %q", i.GetNumber(), i.GetClosedAt().Format(dateForm), i.GetUpdatedAt().Format(dateForm), i.GetTitle())

			full, err := ghcache.IssuesGet(ctx, dv, c, t, org, project, i.GetNumber())
			if err != nil {
				time.Sleep(1)
				full, err = ghcache.IssuesGet(ctx, dv, c, t, org, project, i.GetNumber())
			}
			if err != nil {
				klog.Errorf("failed IssuesGet: %v", err)
				break
			}

			creator := strings.ToLower(full.GetUser().GetLogin())
			closer := strings.ToLower(full.GetClosedBy().GetLogin())
			if len(matchUser) > 0 && !matchUser[creator] && !matchUser[closer] {
				continue
			}

			result = append(result, full)
		}
	}
	klog.Infof("Returning %d issues", len(result))
	return result, nil
}

func issueDate(i *github.Issue) time.Time {
	t := i.GetClosedAt()
	if t.IsZero() {
		t = i.GetUpdatedAt()
	}
	if t.IsZero() {
		t = i.GetCreatedAt()
	}
	return t
}
