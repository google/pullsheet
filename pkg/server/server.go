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

	"github.com/sirupsen/logrus"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/server/job"
)

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
		res, err := s.jobs[0].Render()
		if err != nil {
			logrus.Errorf("rendering job page: %s", err)
		}
		fmt.Fprint(w, res)
	}
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
