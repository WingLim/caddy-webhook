package webhooks

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestGithubHandle(t *testing.T) {
	hc := &HookConf{
		RefName: plumbing.ReferenceName("refs/heads/main"),
	}
	ghHook := Github{}

	for i, test := range []struct {
		body  string
		event string
		code  int
	}{
		{"", "", http.StatusBadRequest},
		{"", "push", http.StatusBadRequest},
		{`{"ref": "refs/heads/main"}`, "push", http.StatusOK},
		{`{"ref": "refs/heads/others}"`, "push", http.StatusBadRequest},
	} {
		req, err := http.NewRequest("POST", "/webhook", bytes.NewBuffer([]byte(test.body)))
		assert.Nil(t, err, fmt.Sprintf("case %d", i))

		if test.event != "" {
			req.Header.Add("X-Github-Event", test.event)
		}

		code, _ := ghHook.Handle(req, hc)

		assert.Equal(t, code, test.code, fmt.Sprintf("case %d", i))
	}
}
