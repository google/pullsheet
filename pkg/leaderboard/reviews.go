package leaderboard

import (
	"github.com/google/pullsheet/pkg/repo"
)

func reviewsChart(reviews []*repo.ReviewSummary) chart {
	uMap := map[string]int{}
	for _, r := range reviews {
		uMap[r.Reviewer]++
	}

	return chart{
		ID:     "reviewCounts",
		Title:  "Top PR Reviewers",
		Metric: "Pull Requests Reviewed",
		Items:  topItems(mapToItems(uMap)),
	}
}

func reviewCommentsChart(reviews []*repo.ReviewSummary) chart {
	uMap := map[string]int{}
	for _, r := range reviews {
		uMap[r.Reviewer] += r.ReviewComments
	}

	return chart{
		ID:     "reviewComments",
		Title:  "Most Nitpicky Reviewers",
		Metric: "Review Comments",
		Items:  topItems(mapToItems(uMap)),
	}
}

func reviewWordsChart(reviews []*repo.ReviewSummary) chart {
	uMap := map[string]int{}
	for _, r := range reviews {
		uMap[r.Reviewer] += r.Words
	}

	return chart{
		ID:     "reviewWords",
		Title:  "Most Helpful Reviewers",
		Metric: "Review Words",
		Items:  topItems(mapToItems(uMap)),
	}
}
