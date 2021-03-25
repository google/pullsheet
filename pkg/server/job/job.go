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
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/leaderboard"
)

type Job struct {
	opts *Opts
	u    *updater
}

// Options related to the Job
type Opts struct {
	Repos []string
	Users []string
	Since time.Time
	Until time.Time
	Title string
}

func New(opts *Opts) *Job {
	return &Job{
		opts: opts,
		u: &updater{
			mu:   &sync.Mutex{},
			data: data{},
		},
	}
}

func (j *Job) Render() (string, error) {
	d := data{
		prs:      j.u.getPRs(),
		reviews:  j.u.getReviews(),
		issues:   j.u.getIssues(),
		comments: j.u.getComments(),
	}

	result, err := leaderboard.Render(j.opts.Title, j.opts.Since, j.opts.Until, j.opts.Users, d.prs, d.reviews, d.issues, d.comments)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (j *Job) Update(ctx context.Context, cl *client.Client) {
	err := j.u.updateData(ctx, cl, j.opts)
	if err != nil {
		logrus.Errorf("Failed to update job: %d", err)
	}
}

func (j *Job) GetOpts() Opts {
	return *j.opts
}
