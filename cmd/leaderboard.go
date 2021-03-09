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
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/leaderboard"
)

// leaderBoardCmd represents the subcommand for `pullsheet leaderboard`
var leaderBoardCmd = &cobra.Command{
	Use:           "leaderboard",
	Short:         "Generate leaderboard data",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLeaderBoard(rootOpts)
	},
}

func init() {
	rootCmd.AddCommand(leaderBoardCmd)
}

func runLeaderBoard(rootOpts *rootOptions) error {
	ctx := context.Background()
	c, err := client.New(ctx, rootOpts.tokenPath)
	if err != nil {
		return err
	}

	prs, err := generatePullData(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.branches, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return err
	}

	reviews, err := generateReviewData(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return err
	}

	issues, err := generateIssueData(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return err
	}

	comments, err := generateCommentsData(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return err
	}

	title := rootOpts.title
	if title == "" {
		title = strings.Join(rootOpts.repos, ", ")
	}

	out, err := leaderboard.Render(title, rootOpts.sinceParsed, rootOpts.untilParsed, rootOpts.users, prs, reviews, issues, comments)
	if err != nil {
		return err
	}

	logrus.Infof("%d bytes of issue-comments output", len(out))
	fmt.Print(out)

	return nil
}
