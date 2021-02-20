package repo

import (
	"net/url"
	"strings"
)

// ParseURL returns the organization and project for a URL or partial path
func ParseURL(rawURL string) (string, string) {
	u, err := url.Parse(rawURL)
	if err == nil {
		p := strings.Split(u.Path, "/")
		if u.Hostname() != "" {
			return p[1], p[2]
		}
		return p[0], p[1]
	}
	// Not a URL
	p := strings.Split(rawURL, "/")
	return p[0], p[1]
}
