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
	"github.com/google/go-github/v29/github"
	"github.com/google/pullsheet/pkg/repo"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
)

const dateForm = "2006-01-02"

var (
	reposFlag       = flag.String("repos", "", "comma-delimited list of repositories. ex: kubernetes/minikube")
	usersFlag       = flag.String("users", "", "comma-delimiited list of users")
	sinceFlag       = flag.String("since", "", "when to query from")
	untilFlag       = flag.String("until", "", "when to query till")
	reviewsFlag     = flag.Bool("reviews", false, "generate data on PR reviews")
	includeFileInfo = flag.Bool("file-metadata", true, "Include file information, such as deltas (slower)")
	tokenPath       = flag.String("token-path", "", "GitHub token path")
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
	out := ""

	if *reviewsFlag {
		out, err = generateReviewData(ctx, c, repos, users, since, until)
	} else {
		out, err = generatePullData(ctx, c, repos, users, since, until, *includeFileInfo)
	}

	if err != nil {
		klog.Exitf("generate failed: %v", err)
	}
	fmt.Print(out)
}

func generateReviewData(ctx context.Context, c *github.Client, repos []string, users []string, since time.Time, until time.Time) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}

func generatePullData(ctx context.Context, c *github.Client, repos []string, users []string, since time.Time, until time.Time, fileInfo bool) (string, error) {
	result := map[*github.PullRequest][]*github.CommitFile{}

	for _, r := range strings.Split(*reposFlag, ",") {
		org, project := repo.ParseURL(r)

		prs, err := repo.ListPulls(ctx, c, org, project, since, until, users)
		if err != nil {
			return "", fmt.Errorf("list: %v", err)
		}

		for _, pr := range prs {
			var files []*github.CommitFile
			var err error

			if fileInfo {
				files, err = repo.FilteredFiles(ctx, c, org, project, pr.GetNumber())
				if err != nil {
					klog.Errorf("unable to get file list for #%d: %v", pr.GetNumber(), err)
				}
			}

			result[pr] = files
		}
	}

	sum, err := repo.PullSummary(result, since, until, *includeFileInfo)
	if err != nil {
		return "", fmt.Errorf("pull summary failed: %v", err)
	}

	return gocsv.MarshalString(&sum)
}
