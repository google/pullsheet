package summary

import (
	"context"
	"fmt"
	"github.com/google/go-github/v33/github"
	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/repo"
	"github.com/sirupsen/logrus"
	"time"
)

func GeneratePullData(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.PRSummary, error) {
	prFiles := map[*github.PullRequest][]github.CommitFile{}

	for _, r := range repos {
		org, project := repo.ParseURL(r)

		prs, err := repo.MergedPulls(ctx, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("list: %v", err)
		}

		for _, pr := range prs {
			files, err := repo.FilteredFiles(ctx, c, pr.GetMergedAt(), org, project, pr.GetNumber())
			if err != nil {
				return nil, fmt.Errorf("filtered files: %v", err)
			}
			logrus.Errorf("%s files: %v", pr, files)
			prFiles[pr] = files
		}
	}

	sum, err := repo.PullSummary(prFiles, since, until)
	if err != nil {
		return nil, fmt.Errorf("pull summary failed: %v", err)
	}

	return sum, nil
}

func GenerateReviewData(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.ReviewSummary, error) {
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

func GenerateIssueData(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.IssueSummary, error) {
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

func GenerateCommentsData(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.CommentSummary, error) {
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
