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

package job

import (
	"context"
	"strings"
	"sync"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/repo"
	"github.com/google/pullsheet/pkg/summary"
)

type updater struct {
	mu   *sync.Mutex
	data data
}

type data struct {
	prs      []*repo.PRSummary
	reviews  []*repo.ReviewSummary
	issues   []*repo.IssueSummary
	comments []*repo.CommentSummary
}

func (u *updater) getPRs() []*repo.PRSummary {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.data.prs
}

func (u *updater) getReviews() []*repo.ReviewSummary {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.data.reviews
}

func (u *updater) getIssues() []*repo.IssueSummary {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.data.issues
}

func (u *updater) getComments() []*repo.CommentSummary {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.data.comments
}

func (u *updater) updateData(ctx context.Context, cl *client.Client, opts *Opts) error {
	// Query data
	prs, err := summary.Pulls(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	reviews, err := summary.Reviews(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	issues, err := summary.Issues(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	comments, err := summary.Comments(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	title := opts.Title
	if title == "" {
		title = strings.Join(opts.Repos, ", ")
	}

	// Update data in Job
	u.mu.Lock()
	defer u.mu.Unlock()

	u.data = data{
		prs:      prs,
		reviews:  reviews,
		issues:   issues,
		comments: comments,
	}
	return nil
}
