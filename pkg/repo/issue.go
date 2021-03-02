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
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/ghcache"
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
func ClosedIssues(ctx context.Context, c *client.Client, org string, project string, since time.Time, until time.Time, users []string) ([]*IssueSummary, error) {
	closed, err := issues(ctx, c, org, project, since, until, users, "closed")
	if err != nil {
		return nil, err
	}

	result := make([]*IssueSummary, 0, len(closed))
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
func issues(ctx context.Context, c *client.Client, org string, project string, since time.Time, until time.Time, users []string, state string) ([]*github.Issue, error) {
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

	logrus.Infof("Gathering issues for %s/%s, users=%q: %+v", org, project, users, opts)
	for page := 1; page != 0; {
		opts.ListOptions.Page = page
		issues, resp, err := c.GitHubClient.Issues.ListByRepo(ctx, org, project, opts)
		if err != nil {
			return result, err
		}

		logrus.Infof("Processing page %d of %s/%s issue results ...", page, org, project)

		page = resp.NextPage
		logrus.Infof("Current issue updated at %s", issues[0].GetUpdatedAt())

		for _, i := range issues {
			if i.IsPullRequest() {
				continue
			}
			if i.GetClosedAt().After(until) {
				logrus.Infof("issue #%d closed at %s", i.GetNumber(), i.GetUpdatedAt())
				continue
			}

			if i.GetUpdatedAt().Before(since) {
				logrus.Infof("Hit issue #%d updated at %s", i.GetNumber(), i.GetUpdatedAt())
				page = 0
				break
			}

			if !i.GetClosedAt().IsZero() && i.GetClosedAt().Before(since) {
				continue
			}

			if state != "" && i.GetState() != state {
				logrus.Infof("Skipping issue #%d (state=%q)", i.GetNumber(), i.GetState())
				continue
			}

			t := issueDate(i)

			logrus.Infof("Fetching #%d (closed %s, updated %s): %q", i.GetNumber(), i.GetClosedAt().Format(dateForm), i.GetUpdatedAt().Format(dateForm), i.GetTitle())

			full, err := ghcache.IssuesGet(ctx, c.Cache, c.GitHubClient, t, org, project, i.GetNumber())
			if err != nil {
				time.Sleep(1 * time.Second)
				full, err = ghcache.IssuesGet(ctx, c.Cache, c.GitHubClient, t, org, project, i.GetNumber())
			}
			if err != nil {
				logrus.Errorf("failed IssuesGet: %v", err)
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

	logrus.Infof("Returning %d issues", len(result))
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
