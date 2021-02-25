package leaderboard

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"
	"time"

	"github.com/google/pullsheet/pkg/repo"
)

type Tmpl struct {
	Title      string
	Categories []Category
}

type Category struct {
	Title  string
	Charts []Chart
}

type Chart struct {
	ID     string
	Object string
	Metric string
	Items  []Item
}

type Item struct {
	Title    string
	Quantity int
}

func Render(repos []string, users []string, since time.Time, until time.Time, prs []*repo.PRSummary) ([]byte, error) {

	p := "./pkg/leaderboard/leaderboard.tmpl"
	outTmpl, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("readfile: %v", err)
	}

	fmap := template.FuncMap{}
	tmpl := template.Must(template.New("leaderboard").Funcs(fmap).Parse(string(outTmpl)))

	ctx := Tmpl{
		Title: "Test",
	}
	var w bytes.Buffer
	err = tmpl.ExecuteTemplate(bufio.NewWriter(&w), "http", ctx)
	return w.Bytes(), err
}
