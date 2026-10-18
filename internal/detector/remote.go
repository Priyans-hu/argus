package detector

import (
	"net/url"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// linkedHosts are the only hosts argus links to, with their platform names.
// Any other host (SSH aliases, GitHub Enterprise, lookalike domains) gets a
// plain owner/name.
var linkedHosts = map[string]string{
	"github.com":    "github",
	"gitlab.com":    "gitlab",
	"bitbucket.org": "bitbucket",
}

// repositoryFromRemote builds the repository info for a git remote URL
// without keeping the URL itself, which can carry credentials or a private
// SSH host alias. It returns nil for local paths and unparseable remotes.
func repositoryFromRemote(remote string) *types.GitRepository {
	host, owner, name, ok := parseRemote(remote)
	if !ok {
		return nil
	}
	repo := &types.GitRepository{Owner: owner, Name: name}
	if platform, ok := linkedHosts[host]; ok {
		repo.Platform = platform
		repo.RemoteURL = "https://" + host + "/" + owner + "/" + name
	}
	return repo
}

// parseRemote splits a remote in scp style (user@host:path, host:path),
// ssh:// or https:// form into host, owner and name. Userinfo, port and a
// trailing .git are dropped; the host is lowercased and loses a leading
// "www.". The path is split at its last slash, so GitLab subgroups stay in
// the owner. Remotes without a host, or with a single-letter host (a Windows
// drive), are local and report false.
func parseRemote(remote string) (host, owner, name string, ok bool) {
	remote = strings.TrimSpace(remote)
	var path string
	if strings.Contains(remote, "://") {
		u, err := url.Parse(remote)
		if err != nil {
			return "", "", "", false
		}
		host, path = u.Hostname(), u.Path
	} else {
		// scp style needs a colon before any slash; anything else is a path
		colon := strings.IndexByte(remote, ':')
		if colon < 0 || strings.Contains(remote[:colon], "/") {
			return "", "", "", false
		}
		host, path = remote[:colon], remote[colon+1:]
		if at := strings.LastIndexByte(host, '@'); at >= 0 {
			host = host[at+1:]
		}
	}

	host = strings.TrimPrefix(strings.ToLower(host), "www.")
	if len(host) < 2 {
		return "", "", "", false
	}

	path = strings.Trim(path, "/")
	path = strings.TrimSuffix(path, ".git")
	slash := strings.LastIndexByte(path, '/')
	if slash < 0 {
		return "", "", "", false
	}
	owner, name = strings.Trim(path[:slash], "/"), path[slash+1:]
	if owner == "" || name == "" {
		return "", "", "", false
	}
	return host, owner, name, true
}
