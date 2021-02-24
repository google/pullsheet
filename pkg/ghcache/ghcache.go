package ghcache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/go-github/v33/github"

	"github.com/peterbourgon/diskv"
	"k8s.io/klog/v2"
)

const (
	keyTime = "2006-01-02T150405"
)

type blob struct {
	PullRequest         github.PullRequest
	CommitFiles         []github.CommitFile
	PullRequestComments []github.PullRequestComment
	IssueComments       []github.IssueComment
	Issue               github.Issue
}

func PullRequestsGet(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) (*github.PullRequest, error) {
	key := fmt.Sprintf("pr-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err != nil {
		klog.V(1).Infof("cache miss for %v: %s", key, err)
		pr, _, err := c.PullRequests.Get(ctx, org, project, num)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		return pr, save(dv, key, blob{PullRequest: *pr})
	}

	klog.V(1).Infof("cache hit: %v", key)
	return &val.PullRequest, nil
}

func PullRequestsListFiles(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) ([]github.CommitFile, error) {
	key := fmt.Sprintf("pr-listfiles-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err != nil {
		klog.V(1).Infof("cache miss for %v: %s", key, err)
		fsp, _, err := c.PullRequests.ListFiles(ctx, org, project, num, &github.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		fs := []github.CommitFile{}
		for _, f := range fsp {
			fs = append(fs, *f)
		}
		return fs, save(dv, key, blob{CommitFiles: fs})
	}

	klog.V(1).Infof("cache hit: %v", key)
	return val.CommitFiles, nil
}

func PullRequestsListComments(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) ([]github.PullRequestComment, error) {
	key := fmt.Sprintf("pr-comments-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err != nil {
		klog.V(1).Infof("cache miss for %v: %s", key, err)
		csp, _, err := c.PullRequests.ListComments(ctx, org, project, num, &github.PullRequestListCommentsOptions{})
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		cs := []github.PullRequestComment{}
		for _, c := range csp {
			cs = append(cs, *c)
		}
		return cs, save(dv, key, blob{PullRequestComments: cs})
	}

	klog.V(1).Infof("cache hit: %v", key)
	return val.PullRequestComments, nil
}

func IssuesGet(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) (*github.Issue, error) {
	key := fmt.Sprintf("issue-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err != nil {
		klog.V(1).Infof("cache miss for %v: %s", key, err)
		i, _, err := c.Issues.Get(ctx, org, project, num)
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		return i, save(dv, key, blob{Issue: *i})
	}

	klog.V(1).Infof("cache hit: %v", key)
	return &val.Issue, nil
}

func IssuesListComments(ctx context.Context, dv *diskv.Diskv, c *github.Client, t time.Time, org string, project string, num int) ([]github.IssueComment, error) {
	key := fmt.Sprintf("issue-comments-%s-%s-%d-%s", org, project, num, t.Format(keyTime))
	val, err := read(dv, key)

	if err != nil {
		klog.V(1).Infof("cache miss for %v: %s", key, err)
		csp, _, err := c.Issues.ListComments(ctx, org, project, num, &github.IssueListCommentsOptions{})
		if err != nil {
			return nil, fmt.Errorf("get: %v", err)
		}
		cs := []github.IssueComment{}
		for _, c := range csp {
			cs = append(cs, *c)
		}
		return cs, save(dv, key, blob{IssueComments: cs})
	}

	klog.V(1).Infof("cache hit: %v", key)
	return val.IssueComments, nil
}

func save(dv *diskv.Diskv, key string, blob blob) error {
	var bs bytes.Buffer
	enc := gob.NewEncoder(&bs)
	err := enc.Encode(blob)
	if err != nil {
		return fmt.Errorf("encode: %v", err)
	}
	return dv.Write(key, bs.Bytes())
}

func read(dv *diskv.Diskv, key string) (blob, error) {
	var bl blob
	val, err := dv.Read(key)
	if err != nil {
		return bl, err
	}

	enc := gob.NewDecoder(bytes.NewBuffer(val))
	err = enc.Decode(&bl)
	return bl, err
}

// New returns a new cache (hardcoded to diskv, for the moment)
func New() (*diskv.Diskv, error) {
	gob.Register(blob{})
	return initialize()
}

// initialize returns an initialized cache
func initialize() (*diskv.Diskv, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("cache dir: %w", err)
	}
	cacheDir := filepath.Join(root, "pullsheet")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	klog.Infof("cache dir is %s", cacheDir)

	return diskv.New(diskv.Options{
		BasePath:     cacheDir,
		CacheSizeMax: 1024 * 1024 * 1024,
	}), nil
}
