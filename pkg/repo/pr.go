// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/ghcache"
)

const dateForm = "2006-01-02"

var (
	ignorePathRe = regexp.MustCompile(`go\.mod|go\.sum|vendor/|third_party|ignore|schemas/v\d|schema/v\d|Gopkg.lock|.DS_Store`)
	truncRe      = regexp.MustCompile(`changelog|CHANGELOG|Gopkg.toml`)
	commentRe    = regexp.MustCompile(`<!--.*?>`)
)

// MergedPulls returns a list of pull requests in a project
func MergedPulls(ctx context.Context, c *client.Client, org string, project string, since time.Time, until time.Time, users []string) ([]*github.PullRequest, error) {
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

	logrus.Infof("Gathering pull requests for %s/%s, users=%q: %+v", org, project, users, opts)
	for page := 1; page != 0; {
		opts.ListOptions.Page = page
		prs, resp, err := c.GitHubClient.PullRequests.List(ctx, org, project, opts)
		if err != nil {
			return result, err
		}

		logrus.Infof("Processing page %d of %s/%s pull request results (looking for %s)...", page, org, project, since)

		page = resp.NextPage
		logrus.Infof("Current PR updated at %s", prs[0].GetUpdatedAt())
		for _, pr := range prs {
			if pr.GetClosedAt().After(until) {
				logrus.Infof("PR#%d closed at %s", pr.GetNumber(), pr.GetUpdatedAt())
				continue
			}

			if pr.GetUpdatedAt().Before(since) {
				logrus.Infof("Hit PR#%d updated at %s", pr.GetNumber(), pr.GetUpdatedAt())
				page = 0
				break
			}

			if !pr.GetClosedAt().IsZero() && pr.GetClosedAt().Before(since) {
				continue
			}

			uname := strings.ToLower(pr.GetUser().GetLogin())
			if len(matchUser) > 0 && !matchUser[uname] {
				continue
			}

			if isBot(pr.GetUser()) {
				continue
			}

			if pr.GetState() != "closed" {
				logrus.Infof("Skipping PR#%d by %s (state=%q)", pr.GetNumber(), pr.GetUser().GetLogin(), pr.GetState())
				continue
			}

			logrus.Infof("Fetching PR #%d by %s (updated %s): %q", pr.GetNumber(), pr.GetUser().GetLogin(), pr.GetUpdatedAt(), pr.GetTitle())
			fullPR, err := ghcache.PullRequestsGet(ctx, c.Cache, c.GitHubClient, pr.GetMergedAt(), org, project, pr.GetNumber())
			if err != nil {
				time.Sleep(1 * time.Second)
				fullPR, err = ghcache.PullRequestsGet(ctx, c.Cache, c.GitHubClient, pr.GetMergedAt(), org, project, pr.GetNumber())
				if err != nil {
					logrus.Errorf("failed PullRequestsGet: %v", err)
					break
				}
			}

			if !fullPR.GetMerged() || fullPR.GetMergeCommitSHA() == "" {
				logrus.Infof("#%d was not merged, skipping", pr.GetNumber())
				continue
			}

			if pr.GetMergedAt().Before(since) {
				logrus.Infof("#%d was merged earlier than %s, skipping", pr.GetNumber(), since)
				continue
			}

			result = append(result, fullPR)
		}
	}
	logrus.Infof("Returning %d pull request results", len(result))
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
func PullSummary(prs map[*github.PullRequest][]*github.CommitFile, since time.Time, until time.Time) ([]*PRSummary, error) {
	sum := []*PRSummary{}
	seen := map[string]bool{}

	for pr, files := range prs {
		if seen[pr.GetHTMLURL()] {
			logrus.Infof("skipping seen issue: %s", pr.GetHTMLURL())
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
			logrus.Infof("skipping %s - closed at %s, after %s", pr.GetHTMLURL(), t, until)
			continue
		}

		if t.Before(since) {
			logrus.Infof("skipping %s - closed at %s, before %s", pr.GetHTMLURL(), t, since)
			continue
		}

		added := 0
		paths := []string{}
		deleted := 0

		for _, f := range files {
			// These files are mostly auto-generated
			if truncRe.MatchString(f.GetFilename()) && f.GetAdditions() > 10 {
				logrus.Infof("truncating %s from %d to %d lines added", f.GetFilename(), f.GetAdditions(), 10)
				added += 10
			} else {
				added += f.GetAdditions()
			}
			deleted += f.GetDeletions()
			paths = append(paths, f.GetFilename())
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
