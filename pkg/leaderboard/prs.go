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
		Title:  "Top Merged PR Creators",
		Metric: "Pull Requests Merged",
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
		Title:  "Top Changers",
		Metric: "LoC",
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
		Title:  "Top Deleters",
		Metric: "LoC",
		Items:  topItems(mapToItems(uMap)),
	}
}
