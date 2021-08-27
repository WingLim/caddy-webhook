package webhooks

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestGogsHandle(t *testing.T) {
	hc := &HookConf{
		RefName: plumbing.ReferenceName("refs/heads/main"),
	}
	ggHook := Gogs{}

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
			req.Header.Add("X-Gogs-Event", test.event)
		}

		code, _ := ggHook.Handle(req, hc)

		assert.Equal(t, code, test.code, fmt.Sprintf("case %d", i))
	}
}

func TestHandleSignature(t *testing.T) {
	req, err := http.NewRequest("", "", bytes.NewBuffer([]byte(signatureBody)))
	assert.Nil(t, err)
	ggHook := Gogs{}

	req.Header.Add("X-Gogs-Signature", "2c1e24122b6697f6683589e3d37e215b53f94a913734ad12bd5033056872d7d7")
	err = ggHook.handleSignature(req, []byte(signatureBody), "48dk7eGJ")
	assert.Nil(t, err)
}

// test data from https://github.com/gogs/gogs/issues/4233
var signatureBody = `{
  "ref": "refs/heads/master",
  "before": "fd9375523bf5fac258594ddfa790c09b1af44951",
  "after": "68885af8c4e894ea05889e854486f022b6cb3fb2",
  "compare_url": "https://git.heavydev.fr/Treeminder/portal/compare/fd9375523bf5fac258594ddfa790c09b1af44951...68885af8c4e894ea05889e854486f022b6cb3fb2",
  "commits": [
    {
      "id": "68885af8c4e894ea05889e854486f022b6cb3fb2",
      "message": "User auth enabled\n",
      "url": "https://git.heavydev.fr/Treeminder/portal/commit/68885af8c4e894ea05889e854486f022b6cb3fb2",
      "author": {
        "name": "Etienne Fachaux",
        "email": "etienne@fachaux.fr",
        "username": "etienne.fachaux"
      },
      "committer": {
        "name": "Etienne Fachaux",
        "email": "etienne@fachaux.fr",
        "username": "etienne.fachaux"
      },
      "timestamp": "2017-03-03T21:20:26Z"
    }
  ],
  "repository": {
    "id": 49,
    "owner": {
      "id": 29,
      "login": "Treeminder",
      "full_name": "Treeminder",
      "email": "",
      "avatar_url": "https://git.heavydev.fr/avatars/29",
      "username": "Treeminder"
    },
    "name": "portal",
    "full_name": "Treeminder/portal",
    "description": "Portail captif permettant aux utilisateurs de Treeminder de s'authentifier sur les points d'accès wifi.",
    "private": true,
    "fork": false,
    "html_url": "https://git.heavydev.fr/Treeminder/portal",
    "ssh_url": "git@git.heavydev.fr:Treeminder/portal.git",
    "clone_url": "https://git.heavydev.fr/Treeminder/portal.git",
    "website": "",
    "stars_count": 0,
    "forks_count": 0,
    "watchers_count": 2,
    "open_issues_count": 0,
    "default_branch": "master",
    "created_at": "2017-03-01T19:41:25Z",
    "updated_at": "2017-03-03T18:57:59Z"
  },
  "pusher": {
    "id": 1,
    "login": "etienne.fachaux",
    "full_name": "Etienne Fachaux",
    "email": "etienne@fachaux.fr",
    "avatar_url": "https://git.heavydev.fr/avatars/1",
    "username": "etienne.fachaux"
  },
  "sender": {
    "id": 1,
    "login": "etienne.fachaux",
    "full_name": "Etienne Fachaux",
    "email": "etienne@fachaux.fr",
    "avatar_url": "https://git.heavydev.fr/avatars/1",
    "username": "etienne.fachaux"
  }
}`
