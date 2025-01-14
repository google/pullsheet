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

package summary

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v33/github"
	"k8s.io/klog/v2"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/repo"
)

// Pulls returns a summary of pull requests for the specified repositories, users, and branches.
func Pulls(ctx context.Context, c *client.Client, repos []string, users []string, branches []string, since time.Time, until time.Time) ([]*repo.PRSummary, error) {
	prFiles := map[*github.PullRequest][]github.CommitFile{}

	for _, r := range repos {
		org, project := repo.ParseURL(r)

		prs, err := repo.MergedPulls(ctx, c, org, project, since, until, users, branches)
		if err != nil {
			return nil, fmt.Errorf("list: %v", err)
		}

		for _, pr := range prs {
			files, err := repo.FilteredFiles(ctx, c, pr.GetMergedAt(), org, project, pr.GetNumber())
			if err != nil {
				return nil, fmt.Errorf("filtered files: %v", err)
			}
			klog.Errorf("%s files: %v", pr, files)

			prFiles[pr] = []github.CommitFile{}

			for _, f := range files {
				prFiles[pr] = append(prFiles[pr], *f)
			}
		}
	}

	sum, err := repo.PullSummary(prFiles, since, until)
	if err != nil {
		return nil, fmt.Errorf("pull summary failed: %v", err)
	}

	return sum, nil
}

func Reviews(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.ReviewSummary, error) {
	rs := []*repo.ReviewSummary{}
	for _, r := range repos {
		org, project := repo.ParseURL(r)
		rrs, err := repo.MergedReviews(ctx, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("merged pulls: %v", err)
		}
		rs = append(rs, rrs...)
	}

	return rs, nil
}

func Issues(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.IssueSummary, error) {
	rs := []*repo.IssueSummary{}
	for _, r := range repos {
		org, project := repo.ParseURL(r)
		rrs, err := repo.ClosedIssues(ctx, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("merged pulls: %v", err)
		}
		rs = append(rs, rrs...)
	}

	return rs, nil
}

func Comments(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.CommentSummary, error) {
	rs := []*repo.CommentSummary{}
	for _, r := range repos {
		org, project := repo.ParseURL(r)
		rrs, err := repo.IssueComments(ctx, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("merged pulls: %v", err)
		}

		rs = append(rs, rrs...)
	}

	return rs, nil
}
