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
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/ghcache"
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
func IssueComments(ctx context.Context, c *client.Client, org string, project string, since time.Time, until time.Time, users []string) ([]*CommentSummary, error) {
	is, err := issues(ctx, c, org, project, since, until, nil, "")
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

		cs, err := ghcache.IssuesListComments(ctx, c.Cache, c.GitHubClient, issueDate(i), org, project, i.GetNumber())
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
