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

package repo

import (
	"context"
	"fmt"

	"github.com/google/go-github/v33/github"
	"github.com/google/pullsheet/pkg/client"
)

// ListRepoNames returns the names of all the repositories of the specified Github organization.
func ListRepoNames(ctx context.Context, c *client.Client, org string) ([]string, error) {

	// Retrieve all the repositories of the specified Github organization
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	var allRepos []string

	for {
		repos, resp, err := c.GitHubClient.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			fmt.Println(err)
			return allRepos, err
		}

		for _, val := range repos {
			allRepos = append(allRepos, org+"/"+val.GetName())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}
