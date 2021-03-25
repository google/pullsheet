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

package ghcache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/google/triage-party/pkg/persist"
	"github.com/sirupsen/logrus"
)

const (
	keyTime = "2006-01-02T150405"
)

type blob struct {
	PullRequest         github.PullRequest
	CommitFiles         []github.CommitFile
	PullRequestComments []github.PullRequestComment
	IssueComments       []github.IssueComment
	Issue               github.Issue
}

func PullRequestsGet(ctx context.Context, p persist.Cacher, c *github.Client, t time.Time, org string, project string, num int) (*github.PullRequest, error) {
	key := fmt.Sprintf("pr-%s-%s-%d", org, project, num)
	val := p.Get(key, t)

	if val != nil {
		return val.GHPullRequest, nil
	}

	if val == nil {
		logrus.Debugf("cache miss for %v", key)
		pr, _, err := c.PullRequests.Get(ctx, org, project, num)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		return pr, p.Set(key, &persist.Blob{GHPullRequest: pr})
	}

	logrus.Debugf("cache hit: %v", key)
	return val.GHPullRequest, nil
}

func PullRequestsListFiles(ctx context.Context, p persist.Cacher, c *github.Client, t time.Time, org string, project string, num int) ([]*github.CommitFile, error) {
	key := fmt.Sprintf("pr-listfiles-%s-%s-%d", org, project, num)
	val := p.Get(key, t)

	if val != nil {
		return val.GHCommitFiles, nil
	}

	logrus.Debugf("cache miss for %v: %s", key)

	opts := &github.ListOptions{PerPage: 100}
	fs := []*github.CommitFile{}

	for {
		fsp, resp, err := c.PullRequests.ListFiles(ctx, org, project, num, opts)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		fs = append(fs, fsp...)

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return fs, p.Set(key, &persist.Blob{GHCommitFiles: fs})

}

func PullRequestsListComments(ctx context.Context, p persist.Cacher, c *github.Client, t time.Time, org string, project string, num int) ([]*github.PullRequestComment, error) {
	key := fmt.Sprintf("pr-comments-%s-%s-%d", org, project, num)
	val := p.Get(key, t)

	if val != nil {
		return val.GHPullRequestComments, nil
	}

	logrus.Debugf("cache miss for %v: %s", key)

	cs := []*github.PullRequestComment{}
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		csp, resp, err := c.PullRequests.ListComments(ctx, org, project, num, opts)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}

		cs = append(cs, csp...)

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return cs, p.Set(key, &persist.Blob{GHPullRequestComments: cs})
}

func IssuesGet(ctx context.Context, p persist.Cacher, c *github.Client, t time.Time, org string, project string, num int) (*github.Issue, error) {
	key := fmt.Sprintf("issue-%s-%s-%d", org, project, num)
	val := p.Get(key, t)

	if val != nil {
		return val.GHIssue, nil
	}

	logrus.Debugf("cache miss for %v: %s", key)

	i, _, err := c.Issues.Get(ctx, org, project, num)
	if err != nil {
		return nil, fmt.Errorf("get: %v", err)
	}

	return i, p.Set(key, &persist.Blob{GHIssue: i})
}

func IssuesListComments(ctx context.Context, p persist.Cacher, c *github.Client, t time.Time, org string, project string, num int) ([]*github.IssueComment, error) {
	key := fmt.Sprintf("issue-comments-%s-%s-%d", org, project, num)
	val := p.Get(key, t)

	if val != nil {
		return val.GHIssueComments, nil
	}

	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	cs := []*github.IssueComment{}
	for {
		csp, resp, err := c.Issues.ListComments(ctx, org, project, num, opts)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}

		cs = append(cs, csp...)

		if resp.NextPage == 0 {
			break
		}

		opts.ListOptions.Page = resp.NextPage
	}

	return cs, p.Set(key, &persist.Blob{GHIssueComments: cs})
}
