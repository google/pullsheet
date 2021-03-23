package server

import (
	"context"
	"fmt"
	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/server/job"
	"net/http"
)

type Server struct {
	ctx context.Context
	cl *client.Client
	jobs []*job.Job
}

func New(ctx context.Context, c *client.Client, initJob *job.Job) *Server {
	jobs := []*job.Job{}
	if initJob != nil {
		jobs = append(jobs, initJob)
		go initJob.Update(ctx, c)
	}

	return &Server{
		ctx: ctx,
		cl: c,
		jobs: jobs,
	}
}

func (s *Server) Root() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _ := s.jobs[0].Render()
		fmt.Fprint(w, res)
	}
}
