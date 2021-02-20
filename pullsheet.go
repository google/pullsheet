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
	includeFileInfo = flag.Bool("file-metadata", true, "Include file information, such as deltas (slower)")
	tokenFlag       = flag.String("token", "", "GitHub token")
)

func main() {

	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Set("alsologtostderr", "true")

	flag.Parse()

	if *reposFlag == "" || *sinceFlag == "" || *tokenFlag == "" {
		fmt.Println("usage: pullsheet --repos <repository> --since 2006-01-02 --token <github token> [--user=<user>]")
		os.Exit(2)
	}

	since, err := time.Parse(dateForm, *sinceFlag)
	if err != nil {
		panic(err)
	}

	until := time.Now()
	if *untilFlag != "" {
		until, err = time.Parse(dateForm, *untilFlag)
		if err != nil {
			panic(err)
		}
	}

	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *tokenFlag}))
	c := github.NewClient(tc)
	result := map[*github.PullRequest][]*github.CommitFile{}
	var users []string

	for _, u := range strings.Split(*usersFlag, ",") {
		if len(u) > 0 {
			users = append(users, u)
		}
	}

	for _, r := range strings.Split(*reposFlag, ",") {
		org, project := repo.ParseURL(r)
		prs, err := repo.ListPulls(ctx, c, org, project, since, until, users)
		if err != nil {
			klog.Exitf("failed: %v", err)
		}

		for _, pr := range prs {
			var files []*github.CommitFile
			var err error

			if *includeFileInfo {
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
		klog.Exitf("sheet creation failed: %v", err)
	}
	out, err := gocsv.MarshalString(&sum)
	if err != nil {
		klog.Exitf("marshal failed: %v", err)
	}
	fmt.Print(out)
}
