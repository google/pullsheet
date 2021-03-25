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

package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/server/job"
	"github.com/google/pullsheet/pkg/server/site"
	"github.com/karrick/tparse"
	"github.com/sirupsen/logrus"
)

const dateForm = "2006-01-02"

type Server struct {
	cl   *client.Client
	jobs []*job.Job
}

func New(ctx context.Context, c *client.Client, initJob *job.Job) *Server {
	jobs := []*job.Job{}
	if initJob != nil {
		jobs = append(jobs, initJob)
		go initJob.Update(ctx, c)
	}

	return &Server{
		cl:   c,
		jobs: jobs,
	}
}

func (s *Server) Root() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/home", http.StatusFound)
	}
}

func (s *Server) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := site.Home(s.jobs)
		if err != nil {
			logrus.Errorf("rendering home page: %d", err)
		}
		fmt.Fprint(w, res)
	}
}

func (s *Server) Job() http.HandlerFunc {
	jobPath := "/job/"
	return func(w http.ResponseWriter, r *http.Request) {
		// Get job number from request URL
		slug := r.URL.Path[len(jobPath):]
		if slug == "" {
			slug = "0"
		}
		idx, err := strconv.Atoi(slug)
		if err != nil {
			logrus.Errorf("getting job index: %d", err)
		}
		if idx >= len(s.jobs) {
			idx = 0
		}

		// Render job from index number
		res, err := s.jobs[idx].Render()
		if err != nil {
			logrus.Errorf("rendering home page: %d", err)
		}
		fmt.Fprint(w, res)
	}
}

func (s *Server) NewJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}

			// Extract form values
			jobName := r.FormValue("jobname")
			repos := r.FormValue("repos")
			users := r.FormValue("users")
			since := r.FormValue("since")
			until := r.FormValue("until")

			sinceParsed, err := tparse.ParseNow(dateForm, since)
			if err != nil {
				logrus.Errorf("Parsing from: %d", err)
			}
			untilParsed, err := tparse.ParseNow(dateForm, until)
			if err != nil {
				logrus.Errorf("Parsing from: %d", err)
			}

			s.AddJob(context.Background(), job.New(&job.Opts{
				Repos: strings.Split(repos, ","),
				Users: strings.Split(users, ","),
				Since: sinceParsed,
				Until: untilParsed,
				Title: jobName,
			}))

			http.Redirect(w, r, fmt.Sprintf("/job/%d", len(s.jobs)-1), http.StatusFound)
		default:
			fmt.Fprintf(w, "Sorry, only POST method is supported.")
		}
	}
}

func (s *Server) AddJob(ctx context.Context, j *job.Job) {
	s.jobs = append(s.jobs, j)
	go j.Update(ctx, s.cl)
}

// Healthz returns a dummy healthz page - it's always happy here!
func (s *Server) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

// Threadz returns a threadz page
func (s *Server) Threadz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("GET %s: %v", r.URL.Path, r.Header)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(stack()); err != nil {
			logrus.Errorf("writing threadz response: %d", err)
		}
	}
}

// stack returns a formatted stack trace of all goroutines
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
func stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
