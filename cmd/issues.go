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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/repo"
)

// issuesCmd represents the subcommand for `pullsheet issues`
var issuesCmd = &cobra.Command{
	Use:           "issues",
	Short:         "Generate data around issues",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runIssues(rootOpts)
	},
}

func init() {
	rootCmd.AddCommand(issuesCmd)
}

func runIssues(rootOpts *rootOptions) error {
	ctx := context.Background()
	c, err := client.New(ctx, rootOpts.tokenPath)
	if err != nil {
		return err
	}

	data, err := generateIssueData(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return err
	}

	out, err := gocsv.MarshalString(&data)
	if err != nil {
		return err
	}

	logrus.Infof("%d bytes of issue output", len(out))
	fmt.Print(out)

	return nil
}

func generateIssueData(ctx context.Context, c *client.Client, repos []string, users []string, since time.Time, until time.Time) ([]*repo.IssueSummary, error) {
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
