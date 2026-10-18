package detector

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// credentialedRemote builds an https remote with userinfo at run time, so no
// credential-shaped URL is ever committed.
func credentialedRemote(secret string) string {
	return "https://" + "deploy:" + secret + "@" + "github.com/o/n.git"
}

func TestRepositoryFromRemote(t *testing.T) {
	github := &types.GitRepository{RemoteURL: "https://github.com/o/n", Owner: "o", Name: "n", Platform: "github"}
	plain := &types.GitRepository{Owner: "o", Name: "n"}

	tests := []struct {
		remote string
		want   *types.GitRepository
	}{
		{"git@github.com:o/n.git", github},
		{"github-personal:o/n.git", plain},
		{"git@github-personal:o/n.git", plain},
		{"ssh://git@github.com:22/o/n.git", github},
		{"https://github.com/o/n", github},
		{"https://www.github.com/o/n.git", github},
		{"HTTPS://GitHub.com/o/n.git/", github},
		{credentialedRemote(strings.Repeat("q7", 4)), github},
		{"https://gitlab.com/g/sub/n.git", &types.GitRepository{
			RemoteURL: "https://gitlab.com/g/sub/n", Owner: "g/sub", Name: "n", Platform: "gitlab"}},
		{"https://bitbucket.org/o/n.git", &types.GitRepository{
			RemoteURL: "https://bitbucket.org/o/n", Owner: "o", Name: "n", Platform: "bitbucket"}},
		{"https://github.example.com/o/n.git", plain},
		{"https://github.com.evil.example/o/n", plain},
		{"https://github.com/n", nil},
		{"/srv/git/x.git", nil},
		{"../x", nil},
		{`C:\repos\x`, nil},
		{"C:/repos/x", nil},
		{"file:///srv/git/x.git", nil},
		{"", nil},
	}
	for _, tt := range tests {
		got := repositoryFromRemote(tt.remote)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("repositoryFromRemote(%q) = %+v, want %+v", tt.remote, got, tt.want)
		}
	}
}
