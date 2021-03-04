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

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/repo"
)

// prsCmd represents the subcommand for `pullsheet prs`
var prsCmd = &cobra.Command{
	Use:           "prs",
	Short:         "Generate data around pull requests",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPRs(rootOpts)
	},
}

func init() {
	rootCmd.AddCommand(prsCmd)
}

func runPRs(rootOpts *rootOptions) error {
	ctx := context.Background()
	c, err := client.New(ctx, rootOpts.tokenPath)
	if err != nil {
		return err
	}

	data, err := generatePullData(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return err
	}

	out, err := gocsv.MarshalString(&data)
	if err != nil {
		return err
	}

	logrus.Infof("%d bytes of prs output", len(out))
	fmt.Print(out)

	return nil
}

func generatePullData(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.PRSummary, error) {
	prFiles := map[*github.PullRequest][]*github.CommitFile{}

	for _, r := range repos {
		org, project := repo.ParseURL(r)

		prs, err := repo.MergedPulls(ctx, c, org, project, since, until, users)
		if err != nil {
			return nil, fmt.Errorf("list: %v", err)
		}

		for _, pr := range prs {
			var files []*github.CommitFile
			var err error

			files, err = repo.FilteredFiles(ctx, c, pr.GetMergedAt(), org, project, pr.GetNumber())
			if err != nil {
				logrus.Errorf("unable to get file list for #%d: %v", pr.GetNumber(), err)
				return nil, err
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
