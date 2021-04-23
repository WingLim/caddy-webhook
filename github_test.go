package caddy_webhook

import (
	"bytes"
	"github.com/go-git/go-git/v5/plumbing"
	"net/http"
	"testing"
)

func TestGithubHandle(t *testing.T) {
	repo := &Repo{
		Branch:  "main",
		Secret:  "supersecret",
		refName: plumbing.ReferenceName("refs/heads/main"),
	}
	ghHook := Github{}

	for i, test := range []struct {
		body   string
		event  string
		secret string
		code   int
	}{
		{"", "", "", http.StatusBadRequest},
		{"", "push", "", http.StatusBadRequest},
		{pushMain, "push", repo.Secret, http.StatusOK},
		{pushMain, "push", "wrongsecret", http.StatusBadRequest},
		{pushOther, "push", repo.Secret, http.StatusBadRequest},
	} {
		req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer([]byte(test.body)))
		if err != nil {
			t.Fatalf("Test %v: Could not create HTTP request: %v", i, err)
		}

		if test.event != "" {
			req.Header.Add("X-Github-Event", test.event)
		}

		code, err := ghHook.Handle(req, repo)

		if code != test.code {
			t.Errorf("Test %d: Expected response code to be %d but was %d", i, test.code, code)
		}
	}
}

var pushMain = `
{
	"ref": "refs/heads/main"
}
`

var pushOther = `
{
	"ref": "refs/heads/some-other-branch"
}
`
