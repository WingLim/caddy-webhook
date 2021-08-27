package webhooks

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestBitbucketHandle(t *testing.T) {
	hc := &HookConf{
		RefName: plumbing.ReferenceName("refs/heads/main"),
	}
	bbHook := Bitbucket{}

	remoteIP := "18.246.31.128"
	remoteIPv6 := "2600:1f18:2146:e306:939f:d1b3:aa36:ac42"
	atlassianIPsMu.Lock()
	atlassianIPs = atlassianIPResponse{
		Items: []atlassianIPRange{
			{
				Network: remoteIP,
				MaskLen: 25,
				CIDR:    remoteIP + "/25",
				Mask:    "255.255.255.128",
			},
			{
				Network: "2600:1f18:2146:e306:939f:d1b3:aa36:ac42",
				MaskLen: 56,
				CIDR:    "2600:1f18:2146:e300::/56",
				Mask:    "ffff:ffff:ffff:ff00::",
			},
		},
		lastUpdated: time.Now(),
	}
	atlassianIPsMu.Unlock()

	for i, test := range []struct {
		ip    string
		body  string
		event string
		code  int
	}{
		{remoteIP, "", "", http.StatusBadRequest},
		{"131.103.20.160", "", "repo:push", http.StatusForbidden},
		{remoteIP, "", "repo:push", http.StatusBadRequest},
		{remoteIP, pushBBBodyValid, "repo:push", http.StatusOK},
		{remoteIPv6, pushBBBodyValid, "repo:push", http.StatusOK},
		{remoteIP, pushBBBodyEmptyBranch, "repo:push", http.StatusBadRequest},
		{remoteIP, pushBBBodyDeleteBranch, "repo:push", http.StatusBadRequest},
	} {
		req, err := http.NewRequest("POST", "", bytes.NewBuffer([]byte(test.body)))
		assert.Nil(t, err, fmt.Sprintf("case %d", i))

		req.RemoteAddr = test.ip

		if test.event != "" {
			req.Header.Add("X-Event-Key", test.event)
		}

		code, _ := bbHook.Handle(req, hc)

		assert.Equal(t, code, test.code, fmt.Sprintf("case %d", i))
	}
}

var pushBBBodyEmptyBranch = `
{
	"push": {
		"changes": [
			{
				"new": {
					"type": "branch",
					"name": ""
				}
			}
		]
	}
}
`

var pushBBBodyValid = `
{
	"push": {
		"changes": [
			{
				"new": {
					"type": "branch",
					"name": "main"
				}
			}
		]
	}
}
`

var pushBBBodyDeleteBranch = `
{
	"push": {
		"changes": [
		]
	}
}
`
