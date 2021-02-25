package leaderboard

import (
	"github.com/google/pullsheet/pkg/repo"
)

func issueCloserChart(is []*repo.IssueSummary) chart {
	uMap := map[string]int{}
	for _, i := range is {
		if i.Author != i.Closer {
			uMap[i.Closer]++
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
