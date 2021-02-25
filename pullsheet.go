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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/google/go-github/v33/github"
	"github.com/google/pullsheet/pkg/ghcache"
	"github.com/google/pullsheet/pkg/leaderboard"
	"github.com/google/pullsheet/pkg/repo"
	"github.com/peterbourgon/diskv"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
)

const dateForm = "2006-01-02"

var (
	reposFlag = flag.String("repos", "", "comma-delimited list of repositories. ex: kubernetes/minikube")
	usersFlag = flag.String("users", "", "comma-delimiited list of users")
	sinceFlag = flag.String("since", "", "when to query from")
	untilFlag = flag.String("until", "", "when to query till")
	modeFlag  = flag.String("mode", "pr", "mode: pr, pr_comment, issue, issue_comment, leaderboard")
	tokenPath = flag.String("token-path", "", "GitHub token path")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("alsologtostderr", "true")

	flag.Parse()

	if *reposFlag == "" || *sinceFlag == "" || *tokenPath == "" {
		fmt.Println("usage: pullsheet --repos <repository> --since 2006-01-02 --token <github token> [--users=<user>]")
		os.Exit(2)
	}

	since, err := time.Parse(dateForm, *sinceFlag)
	if err != nil {
		klog.Exitf("since time parse: %v", err)
	}

	until := time.Now()
	if *untilFlag != "" {
		until, err = time.Parse(dateForm, *untilFlag)
		if err != nil {
			klog.Exitf("until time parse: %v", err)
		}
	}

	ctx := context.Background()
	token, err := ioutil.ReadFile(*tokenPath)
	if err != nil {
		klog.Exitf("token file: %v", err)
	}

	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: strings.TrimSpace(string(token))}))
	c := github.NewClient(tc)
	var users []string

	for _, u := range strings.Split(*usersFlag, ",") {
		if len(u) > 0 {
			users = append(users, u)
		}
	}

	repos := strings.Split(*reposFlag, ",")

	dv, err := ghcache.New()
	if err != nil {
		klog.Exitf("cache: %v", err)
	}

	var out string

	switch *modeFlag {
	case "pr_comment", "pr_comments":
		data, err := generateReviewData(ctx, dv, c, repos, users, since, until)
		if err != nil {
			klog.Exitf("err: %v", err)
		}
		out, err = gocsv.MarshalString(&data)

	case "pr", "prs":
		data, err := generatePullData(ctx, dv, c, repos, users, since, until)
		if err != nil {
			klog.Exitf("err: %v", err)
		}
		out, err = gocsv.MarshalString(&data)

	case "issue", "issues":
		data, err := generateIssueData(ctx, dv, c, repos, users, since, until)
		if err != nil {
			klog.Exitf("err: %v", err)
		}
		out, err = gocsv.MarshalString(&data)

	case "issue_comment", "issue_comments":
		data, err := generateCommentsData(ctx, dv, c, repos, users, since, until)
		if err != nil {
			klog.Exitf("err: %v", err)
		}
		out, err = gocsv.MarshalString(&data)
	case "leaderboard":
		prs, err := generatePullData(ctx, dv, c, repos, users, since, until)
		if err != nil {
			klog.Exitf("pull data: %v", err)
		}
		bs, err := leaderboard.Render(repos, users, since, until, prs)
		if err != nil {
			klog.Exitf("render: %v", err)
		}
		out = string(bs)

	default:
		err = fmt.Errorf("unknown mode: %q", *modeFlag)
	}

	if err != nil {
		klog.Exitf("generate failed: %v", err)
	}
	fmt.Print(out)
}

func generateReviewData(ctx context.Context, dv *diskv.Diskv, c *github.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.ReviewSummary, error) {
	rs := []*repo.ReviewSummary{}
	for _, r := range repos {
		org, project := repo.ParseURL(r)
		rrs, err := repo.MergedReviews(ctx, dv, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("merged pulls: %v", err)
		}
		rs = append(rs, rrs...)
	}
	return rs, nil
}

func generateCommentsData(ctx context.Context, dv *diskv.Diskv, c *github.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.CommentSummary, error) {
	rs := []*repo.CommentSummary{}
	for _, r := range repos {
		org, project := repo.ParseURL(r)
		rrs, err := repo.IssueComments(ctx, dv, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("merged pulls: %v", err)
		}
		rs = append(rs, rrs...)
	}

	return rs, nil
}

func generatePullData(ctx context.Context, dv *diskv.Diskv, c *github.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.PRSummary, error) {
	prFiles := map[*github.PullRequest][]*github.CommitFile{}

	for _, r := range repos {
		org, project := repo.ParseURL(r)

		prs, err := repo.MergedPulls(ctx, dv, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("list: %v", err)
		}

		for _, pr := range prs {
			var files []*github.CommitFile
			var err error

			files, err = repo.FilteredFiles(ctx, dv, c, pr.GetMergedAt(), org, project, pr.GetNumber())
			if err != nil {
				klog.Errorf("unable to get file list for #%d: %v", pr.GetNumber(), err)
			}

			prFiles[pr] = files
		}
	}

	sum, err := repo.PullSummary(prFiles, since, until)
	if err != nil {
		return nil, fmt.Errorf("pull summary failed: %v", err)
	}
	return sum, nil
}

func generateIssueData(ctx context.Context, dv *diskv.Diskv, c *github.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.IssueSummary, error) {
	rs := []*repo.IssueSummary{}
	for _, r := range repos {
		org, project := repo.ParseURL(r)
		rrs, err := repo.ClosedIssues(ctx, dv, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("merged pulls: %v", err)
		}
		rs = append(rs, rrs...)
	}

	return rs, nil
}
