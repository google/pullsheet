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
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"

	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/ghcache"
)

// FilteredFiles returns a list of commit files that matter
func FilteredFiles(ctx context.Context, c *client.Client, t time.Time, org string, project string, num int) ([]*github.CommitFile, error) {
	logrus.Infof("Fetching file list for #%d", num)

	var files []*github.CommitFile
	changed, err := ghcache.PullRequestsListFiles(ctx, c.Cache, c.GitHubClient, t, org, project, num)
	if err != nil {
		return files, err
	}

	for _, cf := range changed {
		if ignorePathRe.MatchString(cf.GetFilename()) {
			logrus.Infof("ignoring %s", cf.GetFilename())
			continue
		}
		files = append(files, &cf)
	}

	return files, err
}

// prType returns what kind of PR it thinks this may be
func prType(files []*github.CommitFile) string {
	result := ""
	for _, cf := range files {
		f := cf.GetFilename()
		ext := strings.TrimLeft(filepath.Ext(f), ".")

		if strings.Contains(filepath.Dir(f), "docs/") || strings.Contains(filepath.Dir(f), "examples/") || strings.Contains(filepath.Dir(f), "site/") {
			if result == "" {
				result = "docs"
			}
			logrus.Infof("%s: %s", f, result)
			continue
		}

		if strings.Contains(f, "test") || strings.Contains(f, "integration") {
			if result == "" {
				result = "tests"
			}
			logrus.Infof("%s: %s", f, result)
			continue
		}

		if ext == "md" && result == "" {
			result = "docs"
		}

		if ext == "go" || ext == "java" || ext == "cpp" || ext == "py" || ext == "c" || ext == "rs" {
			result = "backend"
		}

		if ext == "ts" || ext == "js" || ext == "html" {
			result = "frontend"
		}

		logrus.Infof("%s (ext=%s): %s", f, ext, result)
	}

	if result == "" {
		return "unknown"
	}

	return result
}
