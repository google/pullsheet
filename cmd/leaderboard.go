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
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/pullsheet/pkg/repo"
	"github.com/google/pullsheet/pkg/summary"
	"github.com/karrick/tparse"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/leaderboard"
)

var (
	// leaderBoardCmd represents the subcommand for `pullsheet leaderboard`
	leaderBoardCmd = &cobra.Command{
		Use:           "leaderboard",
		Short:         "Generate leaderboard data",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLeaderBoard(rootOpts)
		},
	}

	disableCaching     bool
	hideCommand        bool
	jsonFiles          []string
	jsonOutput         string
	sinceDisplay       string
	untilDisplay       string
	sinceParsedDisplay time.Time
	untilParsedDisplay time.Time
)

type data struct {
	PRs      []*repo.PRSummary
	Reviews  []*repo.ReviewSummary
	Issues   []*repo.IssueSummary
	Comments []*repo.CommentSummary
}

func init() {
	leaderBoardCmd.Flags().BoolVar(
		&disableCaching,
		"no-caching",
		false,
		"Disable caching on resulting HTML files")

	leaderBoardCmd.Flags().BoolVar(
		&hideCommand,
		"hide-command",
		false,
		"Hide the command-line args in the HTML")

	leaderBoardCmd.Flags().StringSliceVar(
		&jsonFiles,
		"json-files",
		[]string{},
		"List of JSON files to append to the results",
	)

	leaderBoardCmd.Flags().StringVar(
		&jsonOutput,
		"json-output",
		"",
		"Filepath to write the resulting JSON to, will omit if none specified",
	)

	leaderBoardCmd.Flags().StringVar(
		&sinceDisplay,
		"since-display",
		"",
		"This overrides the since date displayed on the leaderboard, primary used if appending past JSON files",
	)

	leaderBoardCmd.Flags().StringVar(
		&untilDisplay,
		"until-display",
		"",
		"This overrides the until date displayed on the leaderboard, primary used if appending past JSON files",
	)

	rootCmd.AddCommand(leaderBoardCmd)
}

func stringToTime(s string, root time.Time) (time.Time, error) {
	if s == "" {
		return root, nil
	}
	parsed, err := tparse.ParseNow(dateForm, s)
	if err != nil {
		klog.Infof("%q not a duration: %v", s, err)
		return time.Time{}, err
	}
	return parsed, nil
}

func runLeaderBoard(rootOpts *rootOptions) error {
	var err error

	sinceParsedDisplay, err = stringToTime(sinceDisplay, rootOpts.sinceParsed)
	if err != nil {
		return err
	}
	untilParsedDisplay, err = stringToTime(untilDisplay, rootOpts.untilParsed)
	if err != nil {
		return err
	}

	d, err := dataFromGitHub()
	if err != nil {
		return err
	}
	d, err = appendJSONFiles(d)
	if err != nil {
		return err
	}

	if err := writeToJSON(d); err != nil {
		return err
	}

	title := rootOpts.title
	if title == "" {
		title = strings.Join(rootOpts.repos, ", ")
	}

	out, err := leaderboard.Render(leaderboard.Options{
		Title:          title,
		Since:          sinceParsedDisplay,
		Until:          untilParsedDisplay,
		DisableCaching: disableCaching,
		HideCommand:    hideCommand,
	}, rootOpts.users, d.PRs, d.Reviews, d.Issues, d.Comments)
	if err != nil {
		return err
	}

	klog.Infof("%d bytes of issue-comments output", len(out))
	fmt.Print(out)

	return nil
}

func dataFromGitHub() (*data, error) {
	ctx := context.Background()
	c, err := client.New(ctx, client.Config{GitHubTokenPath: rootOpts.tokenPath})
	if err != nil {
		return nil, err
	}

	prs, err := summary.Pulls(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.branches, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return nil, err
	}

	reviews, err := summary.Reviews(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return nil, err
	}

	issues, err := summary.Issues(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return nil, err
	}

	comments, err := summary.Comments(ctx, c, rootOpts.repos, rootOpts.users, rootOpts.sinceParsed, rootOpts.untilParsed)
	if err != nil {
		return nil, err
	}

	return &data{prs, reviews, issues, comments}, nil
}

func appendJSONFiles(d *data) (*data, error) {
	for _, file := range jsonFiles {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		var u data
		if err := json.Unmarshal(b, &u); err != nil {
			return nil, err
		}
		d.PRs = append(d.PRs, u.PRs...)
		d.Reviews = append(d.Reviews, u.Reviews...)
		d.Issues = append(d.Issues, u.Issues...)
		d.Comments = append(d.Comments, u.Comments...)
	}
	return d, nil
}

func writeToJSON(d *data) error {
	if jsonOutput == "" {
		return nil
	}
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return os.WriteFile(jsonOutput, b, 0644)
}
