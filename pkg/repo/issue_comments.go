package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/google/pullsheet/pkg/ghcache"
	"github.com/peterbourgon/diskv"
	"k8s.io/klog/v2"
)

// CommentSummary a summary of a users reviews on an issue
type CommentSummary struct {
	URL         string
	Date        string
	Project     string
	Commenter   string
	IssueAuthor string
	IssueState  string
	Comments    int
	Words       int
	Title       string
}

// IssueComments returns a list of issue comment summaries
func IssueComments(ctx context.Context, dv *diskv.Diskv, c *github.Client, org string, project string, since time.Time, until time.Time, users []string) ([]*CommentSummary, error) {
	is, err := issues(ctx, dv, c, org, project, since, until, nil, "")
	if err != nil {
		return nil, fmt.Errorf("issues: %v", err)
	}

	klog.Infof("found %d issues to check comments on", len(is))
	reviews := []*CommentSummary{}

	matchUser := map[string]bool{}
	for _, u := range users {
		matchUser[strings.ToLower(u)] = true
	}

	for _, i := range is {
		if i.IsPullRequest() {
			continue
		}

		// username -> summary
		iMap := map[string]*CommentSummary{}

		cs, err := ghcache.IssuesListComments(ctx, dv, c, issueDate(i), org, project, i.GetNumber())
		if err != nil {
			return nil, err
		}

		for _, c := range cs {
			commenter := c.GetUser().GetLogin()
			if c.CreatedAt.After(until) {
				continue
			}

			if c.CreatedAt.Before(since) {
				continue
			}

			if len(matchUser) > 0 && !matchUser[strings.ToLower(commenter)] {
				continue
			}

			if commenter == i.GetUser().GetLogin() {
				continue
			}

			if isBot(c.GetUser()) {
				continue
			}

			body := strings.TrimSpace(i.GetBody())
			if (strings.HasPrefix(body, "/") || strings.HasPrefix(body, "cc")) && len(body) < 64 {
				klog.Infof("ignoring tag comment: %q", body)
				continue
			}

			wordCount := wordCount(c.GetBody())

			if iMap[commenter] == nil {
				iMap[commenter] = &CommentSummary{
					URL:         i.GetHTMLURL(),
					IssueAuthor: i.GetUser().GetLogin(),
					IssueState:  i.GetState(),
					Commenter:   commenter,
					Project:     project,
					Title:       strings.TrimSpace(i.GetTitle()),
				}
			}

			iMap[commenter].Comments++
			iMap[commenter].Date = c.CreatedAt.Format(dateForm)
			iMap[commenter].Words += wordCount
			klog.Infof("%d word comment by %s: %q for %s/%s #%d", wordCount, commenter, strings.TrimSpace(c.GetBody()), org, project, i.GetNumber())
		}

		for _, rs := range iMap {
			reviews = append(reviews, rs)
		}
	}

	return reviews, err
}
