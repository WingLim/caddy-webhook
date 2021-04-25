package webhooks

import (
	"bytes"
	"fmt"
	"github.com/alecthomas/assert"
	"github.com/go-git/go-git/v5/plumbing"
	"net/http"
	"testing"
)

func TestGiteeHandle(t *testing.T) {
	hc := &HookConf{
		RefName: plumbing.ReferenceName("refs/heads/main"),
	}
	glHook := Gitee{}

	for i, test := range []struct {
		body  string
		event string
		code  int
	}{
		{"", "", http.StatusBadRequest},
		{"", "Push Hook", http.StatusBadRequest},
		{`{"ref": "refs/heads/main"}`, "Push Hook", http.StatusOK},
		{`{"ref": "refs/heads/others}"`, "Push Hook", http.StatusBadRequest},
	} {
		req, err := http.NewRequest("POST", "", bytes.NewBuffer([]byte(test.body)))
		assert.Nil(t, err, fmt.Sprintf("case %d", i))

		if test.event != "" {
			req.Header.Add("X-Gitee-Event", test.event)
		}

		code, _ := glHook.Handle(req, hc)

		assert.Equal(t, code, test.code, fmt.Sprintf("case %d", i))
	}
}
