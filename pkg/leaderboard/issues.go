package leaderboard

import (
	"github.com/google/pullsheet/pkg/repo"
)

func issueCloserChart(is []*repo.IssueSummary) chart {
	uMap := map[string]int{}
	for _, i := range is {
		uMap[i.Closer]++
	}

	return chart{
		ID:     "issueCloser",
		Title:  "Top Issue Closers",
		Metric: "Issues Closed",
		Items:  topItems(mapToItems(uMap)),
	}
}

func commentWordsChart(cs []*repo.CommentSummary) chart {
	uMap := map[string]int{}
	for _, c := range cs {
		uMap[c.Commenter] += c.Words
	}

	return chart{
		ID:     "commentWords",
		Title:  "Most Helpful Commenter",
		Metric: "Words",
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
		Title:  "Most PR comments",
		Metric: "PR Comments",
		Items:  topItems(mapToItems(uMap)),
	}
}
