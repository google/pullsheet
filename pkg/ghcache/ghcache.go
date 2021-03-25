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
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/peterbourgon/diskv"
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

func PullRequestsGet(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) (*github.PullRequest, error) {
	key := fmt.Sprintf("pr-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)
	if err != nil {
		logrus.Debugf("cache miss for %v: %s", key, err)
		pr, _, err := c.PullRequests.Get(ctx, org, project, num)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		return pr, save(dv, key, &blob{PullRequest: *pr})
	}

	logrus.Debugf("cache hit: %v", key)
	return &val.PullRequest, nil
}

func PullRequestsListFiles(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) ([]github.CommitFile, error) {
	key := fmt.Sprintf("pr-listfiles-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err == nil {
		logrus.Debugf("cache hit: %v", key)
		return val.CommitFiles, nil
	}

	opts := &github.ListOptions{PerPage: 100}
	fs := []github.CommitFile{}

	for {
		logrus.Debugf("cache miss for %v: %s", key, err)
		fsp, resp, err := c.PullRequests.ListFiles(ctx, org, project, num, opts)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}

		for _, f := range fsp {
			fs = append(fs, *f)
		}

		if resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return fs, save(dv, key, &blob{CommitFiles: fs})

}

func PullRequestsListComments(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) ([]github.PullRequestComment, error) {
	key := fmt.Sprintf("pr-comments-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err == nil {
		logrus.Debugf("cache hit: %v", key)
		return val.PullRequestComments, nil
	}

	logrus.Debugf("cache miss for %v: %s", key, err)

	cs := []github.PullRequestComment{}
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		csp, resp, err := c.PullRequests.ListComments(ctx, org, project, num, opts)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}

		for _, c := range csp {
			cs = append(cs, *c)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return cs, save(dv, key, &blob{PullRequestComments: cs})
}

func IssuesGet(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) (*github.Issue, error) {
	key := fmt.Sprintf("issue-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)
	if err != nil {
		logrus.Debugf("cache miss for %v: %s", key, err)
		i, _, err := c.Issues.Get(ctx, org, project, num)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		return i, save(dv, key, &blob{Issue: *i})
	}

	logrus.Debugf("cache hit: %v", key)
	return &val.Issue, nil
}

func IssuesListComments(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) ([]github.IssueComment, error) {
	key := fmt.Sprintf("issue-comments-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err == nil {
		logrus.Debugf("cache hit: %v", key)
		return val.IssueComments, nil
	}

	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	cs := []github.IssueComment{}
	for {
		logrus.Debugf("cache miss for %v: %s", key, err)
		csp, resp, err := c.Issues.ListComments(ctx, org, project, num, opts)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}

		for _, c := range csp {
			cs = append(cs, *c)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return cs, save(dv, key, &blob{IssueComments: cs})
}

func save(dv *diskv.Diskv, key string, blob *blob) error {
	var bs bytes.Buffer
	enc := gob.NewEncoder(&bs)
	err := enc.Encode(blob)
	if err != nil {
		return fmt.Errorf("encode: %v", err)
	}
	return dv.Write(key, bs.Bytes())
}

func read(dv *diskv.Diskv, key string) (blob, error) {
	var bl blob
	val, err := dv.Read(key)
	if err != nil {
		return bl, err
	}

	enc := gob.NewDecoder(bytes.NewBuffer(val))
	err = enc.Decode(&bl)
	return bl, err
}

// New returns a new cache (hardcoded to diskv, for the moment)
func New() (*diskv.Diskv, error) {
	gob.Register(blob{})
	return initialize()
}

// initialize returns an initialized cache
func initialize() (*diskv.Diskv, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("cache dir: %w", err)
	}
	cacheDir := filepath.Join(root, "pullsheet")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	logrus.Infof("cache dir is %s", cacheDir)

	return diskv.New(diskv.Options{
		BasePath:     cacheDir,
		CacheSizeMax: 1024 * 1024 * 1024,
	}), nil
}
