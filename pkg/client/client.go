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
	"io/ioutil"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/peterbourgon/diskv"
	"golang.org/x/oauth2"

	"github.com/google/pullsheet/pkg/ghcache"
)

type Client struct {
	Cache        *diskv.Diskv
	GitHubClient *github.Client
}

func New(ctx context.Context, tokenPath string) (*Client, error) {
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: strings.TrimSpace(string(token))}))
	c := github.NewClient(tc)

	dv, err := ghcache.New()
	if err != nil {
		return nil, err
	}

	return &Client{
		Cache:        dv,
		GitHubClient: c,
	}, nil
}
