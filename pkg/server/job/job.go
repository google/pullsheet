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

	"k8s.io/klog/v2"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/leaderboard"
)

// Job represents a job to be run by the server
type Job struct {
	opts *Opts
	u    *updater
}

// Opts Options related to the Job
type Opts struct {
	Repos          []string  // Repos to query
	Branches       []string  // Branches to query
	Users          []string  // Users to query
	Since          time.Time // Since when to query
	Until          time.Time // Until when to query
	Title          string    // Title of the leaderboard
	DisableCaching bool      // Disable caching
}

// New creates a new Job
func New(opts *Opts) *Job {
	return &Job{
		opts: opts,
		u: &updater{
			mu:   &sync.Mutex{},
			data: data{},
		},
	}
}

// Render renders the leaderboard
func (j *Job) Render() (string, error) {
	d := data{
		prs:      j.u.getPRs(),
		reviews:  j.u.getReviews(),
		issues:   j.u.getIssues(),
		comments: j.u.getComments(),
	}

	result, err := leaderboard.Render(leaderboard.Options{
		Title:          j.opts.Title,
		Since:          j.opts.Since,
		Until:          j.opts.Until,
		DisableCaching: j.opts.DisableCaching,
	}, j.opts.Users, d.prs, d.reviews, d.issues, d.comments)
	if err != nil {
		return "", err
	}

	return result, nil
}

// Update updates the Job
func (j *Job) Update(ctx context.Context, cl *client.Client) {
	err := j.u.updateData(ctx, cl, j.opts)
	if err != nil {
		klog.Errorf("Failed to update job: %d", err)
	}
}

// GetOpts returns the options for the Job
func (j *Job) GetOpts() Opts {
	return *j.opts
}
