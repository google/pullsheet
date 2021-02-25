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

func deleteChart(prs []*repo.PRSummary) chart {
	uMap := map[string]int{}
	for _, pr := range prs {
		uMap[pr.User] += pr.Deleted
	}

	return chart{
		ID:     "prDeleters",
		Title:  "Code Slayers",
		Metric: "Lines of code (deleted)",
		Items:  topItems(mapToItems(uMap)),
	}
}
