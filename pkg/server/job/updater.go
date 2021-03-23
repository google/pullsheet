package job

import (
	"context"
	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/repo"
	"github.com/google/pullsheet/pkg/summary"
	"strings"
	"sync"
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

func (u *updater) updateData(ctx context.Context, cl *client.Client, opts Opts) error {
	// Query data
	prs, err := summary.GeneratePullData(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	reviews, err := summary.GenerateReviewData(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	issues, err := summary.GenerateIssueData(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	comments, err := summary.GenerateCommentsData(ctx, cl, opts.Repos, opts.Users, opts.Since, opts.Until)
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
		prs: prs,
		reviews: reviews,
		issues: issues,
		comments: comments,
	}
	return nil
}