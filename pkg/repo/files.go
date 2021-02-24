package repo

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/google/pullsheet/pkg/ghcache"
	"github.com/peterbourgon/diskv"
	"k8s.io/klog/v2"
)

func FilteredFiles(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) ([]*github.CommitFile, error) {
	klog.Infof("Fetching file list for #%d", num)

	var files []*github.CommitFile
	changed, err := ghcache.PullRequestsListFiles(ctx, dv, c, t, org, project, num)
	if err != nil {
		return files, err
	}

	for _, cf := range changed {
		if ignorePathRe.MatchString(cf.GetFilename()) {
			klog.Infof("ignoring %s", cf.GetFilename())
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
			klog.Infof("%s: %s", f, result)
			continue
		}

		if strings.Contains(f, "test") || strings.Contains(f, "integration") {
			if result == "" {
				result = "tests"
			}
			klog.Infof("%s: %s", f, result)
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

		klog.Infof("%s (ext=%s): %s", f, ext, result)
	}

	if result == "" {
		return "unknown"
	}
	return result
}
