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

func mergeChart(prs []*repo.PRSummary) chart {
	uMap := map[string]int{}
	for _, pr := range prs {
		uMap[pr.User]++
	}

	return chart{
		ID:     "prCounts",
		Title:  "Most Active",
		Metric: "# of Pull Requests Merged",
		Items:  topItems(mapToItems(uMap)),
	}
}

func deltaChart(prs []*repo.PRSummary) chart {
	uMap := map[string]int{}
	for _, pr := range prs {
		uMap[pr.User] += pr.Delta
	}

	return chart{
		ID:     "prDeltas",
		Title:  "Big Movers",
		Metric: "Lines of code (delta)",
		Items:  topItems(mapToItems(uMap)),
	}
}

func sizeChart(prs []*repo.PRSummary) chart {
	sz := map[string][]int{}
	for _, pr := range prs {
		sz[pr.User] = append(sz[pr.User], pr.Delta-pr.Deleted)
	}

	uMap := map[string]int{}
	for u, deltas := range sz {
		sum := 0
		for _, delta := range deltas {
			sum += delta
		}

		uMap[u] = sum / len(deltas)
	}

	return chart{
		ID:     "prSize",
		Title:  "Most difficult to review",
		Metric: "Average PR size (added+changed)",
		Items:  topItems(mapToItems(uMap)),
	}
}
