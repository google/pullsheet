package site

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/google/pullsheet/pkg/server/job"
)

//go:embed template/*
var content embed.FS

type jobData struct {
	Title string
}

// Home	render the home page
func Home(jobs []*job.Job) (string, error) {
	jData := []jobData{}
	for _, job := range jobs {
		jData = append(jData, jobData{
			Title: job.GetOpts().Title,
		})
	}

	data := struct {
		Jobs []jobData
	}{
		Jobs: jData,
	}

	t, err := template.ParseFS(content, "template/home.html")
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	out := tpl.String()
	return out, nil
}
