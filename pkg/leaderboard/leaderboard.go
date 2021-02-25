package leaderboard

import (
	"bytes"
	"fmt"
	"path"
	"sort"
	"text/template"
	"time"

	"github.com/google/pullsheet/pkg/repo"
)

const dateForm = "2006-01-02"

// TopX is how many items to include in graphs
var TopX = 15

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
func Render(repos []string, users []string, since time.Time, until time.Time, prs []*repo.PRSummary, reviews []*repo.ReviewSummary, issues []*repo.IssueSummary, comments []*repo.CommentSummary) (string, error) {
	files := []string{"pkg/leaderboard/leaderboard.tmpl"}
	name := path.Base(files[0])
	funcMap := template.FuncMap{}
	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		return "", fmt.Errorf("parsefiles: %v", err)
	}

	data := struct {
		Title      string
		Categories []category
	}{
		Title: fmt.Sprintf("From %s to %s", since.Format(dateForm), until.Format(dateForm)),
		Categories: []category{
			{
				Title: "Reviewers",
				Charts: []chart{
					reviewCommentsChart(reviews),
					reviewWordsChart(reviews),
					reviewsChart(reviews),
				},
			},
			{
				Title: "Pull Requests",
				Charts: []chart{
					mergeChart(prs),
					deltaChart(prs),
					deleteChart(prs),
				},
			},
			{
				Title: "Issues",
				Charts: []chart{
					issueCloserChart(issues),
					commentWordsChart(comments),
					commentsChart(comments),
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
