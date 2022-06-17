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

package leaderboard

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/google/pullsheet/pkg/repo"
)

const dateForm = "2006-01-02"

// TopX is how many items to include in graphs
var TopX = 15

// Options to use for rendering the leaderboard
type Options struct {
	Title          string
	Since          time.Time
	Until          time.Time
	DisableCaching bool
	HideCommand    bool
}

type category struct {
	Title  string
	Charts []chart
}

type chart struct {
	ID     string
	Title  string
	Object string
	Metric string
	Items  []item
}

type item struct {
	Name  string
	Count int
}

// Render returns an HTML formatted leaderboard page
func Render(options Options, users []string, prs []*repo.PRSummary, reviews []*repo.ReviewSummary, issues []*repo.IssueSummary, comments []*repo.CommentSummary) (string, error) {
	funcMap := template.FuncMap{}
	tmpl, err := template.New("LeaderBoard").Funcs(funcMap).Parse(leaderboardTmpl)
	if err != nil {
		return "", fmt.Errorf("parsefiles: %v", err)
	}

	data := struct {
		Title          string
		From           string
		Until          string
		DisableCaching bool
		Command        string
		HideCommand    bool
		Categories     []category
	}{
		Title:          options.Title,
		From:           options.Since.Format(dateForm),
		Until:          options.Until.Format(dateForm),
		DisableCaching: options.DisableCaching,
		Command:        filepath.Base(os.Args[0]) + " " + strings.Join(os.Args[1:], " "),
		HideCommand:    options.HideCommand,
		Categories: []category{
			{
				Title: "Reviewers",
				Charts: []chart{
					reviewsChart(reviews, users),
					reviewWordsChart(reviews, users),
					reviewCommentsChart(reviews, users),
				},
			},
			{
				Title: "Pull Requests",
				Charts: []chart{
					mergeChart(prs, users),
					deltaChart(prs, users),
					sizeChart(prs, users),
				},
			},
			{
				Title: "Issues",
				Charts: []chart{
					commentsChart(comments, users),
					commentWordsChart(comments, users),
					issueCloserChart(issues, users),
				},
			},
		},
	}

	var tpl bytes.Buffer
	if err = tmpl.Execute(&tpl, data); err != nil {
		return "", fmt.Errorf("execute: %w", err)
	}

	out := tpl.String()
	return out, nil
}

func topItems(items []item) []item {
	sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })

	if len(items) > TopX {
		items = items[:TopX]
	}
	return items
}

func mapToItems(m map[string]int) []item {
	items := []item{}
	for u, count := range m {
		items = append(items, item{Name: u, Count: count})
	}
	return items
}
