package repo

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/blevesearch/segment"
	"github.com/google/go-github/v33/github"
	"github.com/google/pullsheet/pkg/ghcache"
	"github.com/peterbourgon/diskv"
	"k8s.io/klog/v2"
)

var (
	notSegmentRe = regexp.MustCompile(`[/-_]+`)
)

// ReviewSummary a summary of a users reviews on a PR
type ReviewSummary struct {
	URL            string
	Date           string
	Project        string
	Reviewer       string
	PRAuthor       string
	PRComments     int
	ReviewComments int
	Words          int
	Title          string
}

type comment struct {
	Author    string
	Body      string
	Review    bool
	CreatedAt time.Time
}

// MergedReviews returns a list of pull requests in a project (merged only)
func MergedReviews(ctx context.Context, dv *diskv.Diskv, c *github.Client, org string, project string, since time.Time, until time.Time, users []string) ([]*ReviewSummary, error) {
	prs, err := MergedPulls(ctx, dv, c, org, project, since, until, nil)
	if err != nil {
		return nil, fmt.Errorf("pulls: %v", err)
	}

	klog.Infof("found %d PR's to check reviews for", len(prs))
	reviews := []*ReviewSummary{}

	matchUser := map[string]bool{}
	for _, u := range users {
		matchUser[strings.ToLower(u)] = true
	}

	for _, pr := range prs {
		// username -> summary
		prMap := map[string]*ReviewSummary{}
		comments := []comment{}

		// There is wickedness in the GitHub API: PR comments are available via the Issues API, and PR *review* comments are available via the PullRequests API
		cs, err := ghcache.PullRequestsListComments(ctx, dv, c, pr.GetMergedAt(), org, project, pr.GetNumber())
		if err != nil {
			return nil, err
		}

		for _, c := range cs {
			if isBot(c.GetUser()) {
				continue
			}
			body := strings.TrimSpace(c.GetBody())
			comments = append(comments, comment{Author: c.GetUser().GetLogin(), Body: body, CreatedAt: c.GetCreatedAt(), Review: true})
		}

		is, err := ghcache.IssuesListComments(ctx, dv, c, pr.GetMergedAt(), org, project, pr.GetNumber())
		if err != nil {
			return nil, err
		}
		for _, i := range is {
			if isBot(i.GetUser()) {
				continue
			}
			body := strings.TrimSpace(i.GetBody())
			if (strings.HasPrefix(body, "/") || strings.HasPrefix(body, "cc")) && len(body) < 64 {
				klog.Infof("ignoring tag comment: %q", body)
				continue
			}

			klog.Infof("%s on #%d is not a bot: %q", i.GetUser().GetLogin(), pr.GetNumber(), body)
			comments = append(comments, comment{Author: i.GetUser().GetLogin(), Body: body, CreatedAt: i.GetCreatedAt(), Review: false})
		}

		for _, c := range comments {
			if len(matchUser) > 0 && !matchUser[strings.ToLower(c.Author)] {
				continue
			}

			if c.Author == pr.GetUser().GetLogin() {
				continue
			}

			wordCount := wordCount(c.Body)

			if prMap[c.Author] == nil {
				prMap[c.Author] = &ReviewSummary{
					URL:      pr.GetHTMLURL(),
					PRAuthor: pr.GetUser().GetLogin(),
					Reviewer: c.Author,
					Project:  project,
					Title:    strings.TrimSpace(pr.GetTitle()),
				}
			}

			if c.Review {
				prMap[c.Author].ReviewComments++
			} else {
				prMap[c.Author].PRComments++
			}

			prMap[c.Author].Date = c.CreatedAt.Format(dateForm)
			prMap[c.Author].Words += wordCount
			klog.Infof("%d word comment by %s: %q for %s/%s #%d", wordCount, c.Author, strings.TrimSpace(c.Body), org, project, pr.GetNumber())
		}

		for _, rs := range prMap {
			reviews = append(reviews, rs)
		}
	}

	return reviews, err
}

// wordCount counts words in a string, irrespective of language
func wordCount(s string) int {
	// Don't count certain items, like / or - as word segments
	s = notSegmentRe.ReplaceAllString(s, "")

	words := 0
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(segment.SplitWords)
	for scanner.Scan() {
		if !unicode.IsLetter(rune(scanner.Bytes()[0])) {
			continue
		}
		words++
	}
	return words
}

func isBot(u *github.User) bool {
	if u.GetType() == "bot" {
		return true
	}

	if strings.Contains(u.GetBio(), "stale issues") {
		return true
	}

	if strings.HasSuffix(u.GetLogin(), "-bot") || strings.HasSuffix(u.GetLogin(), "-robot") || strings.HasSuffix(u.GetLogin(), "_bot") || strings.HasSuffix(u.GetLogin(), "_robot") {
		return true
	}

	if strings.HasPrefix(u.GetLogin(), "codecov") || strings.HasPrefix(u.GetLogin(), "Travis") {
		return true
	}

	return false
}
