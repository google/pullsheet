package job

import (
	"context"
	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/leaderboard"
	"sync"
	"time"
)

type Job struct {
	opts Opts
	u    updater
}

// Options related to the Job
type Opts struct {
	Repos []string
	Users []string
	Since time.Time
	Until time.Time
	Title string
}

func New(opts Opts) *Job {
	return &Job{
		opts: opts,
		u: updater{
			mu: &sync.Mutex{},
			data: data{},
		},
	}
}

func (j *Job) Render() (string, error) {
	d := data {
		prs: j.u.getPRs(),
		reviews: j.u.getReviews(),
		issues: j.u.getIssues(),
		comments: j.u.getComments(),
	}

	result, err := leaderboard.Render(j.opts.Title, j.opts.Since, j.opts.Until, j.opts.Users, d.prs, d.reviews, d.issues, d.comments)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (j *Job) Update(ctx context.Context, cl *client.Client) {
	j.u.updateData(ctx, cl, j.opts)
}
