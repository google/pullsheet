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
	"strings"

	"github.com/google/pullsheet/pkg/repo"
)

func issueCloserChart(is []*repo.IssueSummary) chart {
	uMap := map[string]int{}
	for _, i := range is {
		if i.Author != i.Closer {
			if !strings.HasSuffix(i.Closer, "bot") {
				uMap[i.Closer]++
			}
		}
	}

	return chart{
		ID:     "issueCloser",
		Title:  "Top Closers",
		Metric: "# of issues closed (excludes authored)",
		Items:  topItems(mapToItems(uMap)),
	}
}

func commentWordsChart(cs []*repo.CommentSummary) chart {
	uMap := map[string]int{}
	for _, c := range cs {
		if c.IssueAuthor != c.Commenter {
			uMap[c.Commenter] += c.Words
		}
	}

	return chart{
		ID:     "commentWords",
		Title:  "Most Helpful",
		Metric: "# of words (excludes authored)",
		Items:  topItems(mapToItems(uMap)),
	}
}

func commentsChart(cs []*repo.CommentSummary) chart {
	uMap := map[string]int{}
	for _, c := range cs {
		uMap[c.Commenter] += c.Comments
	}

	return chart{
		ID:     "comments",
		Title:  "Most Active",
		Metric: "# of comments",
		Items:  topItems(mapToItems(uMap)),
	}
}
