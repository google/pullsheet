package leaderboard

import (
	"bytes"
	"fmt"
	"path"
	"text/template"
	"time"

	"github.com/google/pullsheet/pkg/repo"
)

type category struct {
	Title  string
	Charts []chart
}

type chart struct {
	ID     string
	Object string
	Metric string
	Items  []item
}

type item struct {
	Title    string
	Quantity int
}

// Render returns an HTML formatted leaderboard page
func Render(repos []string, users []string, since time.Time, until time.Time, prs []*repo.PRSummary) (string, error) {
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
		Title: "Test",
		Categories: []category{
			{Title: "Pull Requests"},
		},
	}

	var tpl bytes.Buffer
	if err = tmpl.Execute(&tpl, data); err != nil {
		return "", fmt.Errorf("execute: %w", err)
	}

	out := tpl.String()
	return out, nil
}
