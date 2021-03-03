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
	"github.com/google/pullsheet/pkg/repo"
)

func reviewsChart(reviews []*repo.ReviewSummary, _ []string) chart {
	uMap := map[string]int{}
	for _, r := range reviews {
		uMap[r.Reviewer]++
	}

	return chart{
		ID:     "reviewCounts",
		Title:  "Most Influential",
		Metric: "# of Merged PRs reviewed",
		Items:  topItems(mapToItems(uMap)),
	}
}

func reviewCommentsChart(reviews []*repo.ReviewSummary, _ []string) chart {
	uMap := map[string]int{}
	for _, r := range reviews {
		uMap[r.Reviewer] += r.ReviewComments
	}

	return chart{
		ID:     "reviewComments",
		Title:  "Most Demanding",
		Metric: "# of Review Comments in merged PRs",
		Items:  topItems(mapToItems(uMap)),
	}
}

func reviewWordsChart(reviews []*repo.ReviewSummary, _ []string) chart {
	uMap := map[string]int{}
	for _, r := range reviews {
		uMap[r.Reviewer] += r.Words
	}

	return chart{
		ID:     "reviewWords",
		Title:  "Most Helpful",
		Metric: "# of words written in merged PRs",
		Items:  topItems(mapToItems(uMap)),
	}
}
