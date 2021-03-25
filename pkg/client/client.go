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

package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"

	"github.com/google/triage-party/pkg/persist"
)

type Client struct {
	Cache        persist.Cacher
	GitHubClient *github.Client
}

type Config struct {
	GitHubTokenPath string
	GitHubToken     string
	PersistBackend  string
	PersistPath     string
}

func New(ctx context.Context, c Config) (*Client, error) {
	if c.PersistBackend == "" {
		c.PersistBackend = os.Getenv("PERSIST_BACKEND")
	}

	if c.PersistPath == "" {
		c.PersistPath = os.Getenv("PERSIST_PATH")
	}

	if c.GitHubToken == "" {
		c.GitHubToken = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	}

	if c.GitHubToken == "" {
		bs, err := ioutil.ReadFile(c.GitHubTokenPath)
		if err != nil {
			return nil, err
		}
		c.GitHubToken = strings.TrimSpace(string(bs))
	}

	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: c.GitHubToken}))
	gc := github.NewClient(tc)

	p, err := persist.FromEnv("pullsheet", c.PersistBackend, c.PersistPath)
	if err != nil {
		return nil, fmt.Errorf("persist fromenv: %v", err)
	}

	if err := p.Initialize(); err != nil {
		return nil, fmt.Errorf("persist init: %v", err)
	}

	return &Client{
		Cache:        p,
		GitHubClient: gc,
	}, nil
}
