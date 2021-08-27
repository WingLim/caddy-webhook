package webhooks

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
)

type Github struct {
}

type ghPush struct {
	Ref string `json:"ref"`
}

type ghRelease struct {
	Action  string `json:"action"`
	Release struct {
		TagName string `json:"tag_name"`
	} `json:"release"`
}

func (g Github) Handle(r *http.Request, hc *HookConf) (int, error) {
	body, err := ioutil.ReadAll(r.Body)
	err = g.handleSignature(r, body, hc.Secret)
	if err != nil {
		return http.StatusBadRequest, err
	}

	event := r.Header.Get("X-Github-Event")
	if event == "" {
		return http.StatusBadRequest, fmt.Errorf("header 'X-Github-Event' missing")
	}

	switch event {
	case "ping":
	case "push":
		err = g.handlePush(body, hc)
		if err != nil {
			return http.StatusBadRequest, err
		}
	case "release":
		err = g.handleRelease(body, hc)
		if err != nil {
			return http.StatusBadRequest, err
		}
	default:
		return http.StatusBadRequest, fmt.Errorf("cannot handle %q event", event)
	}

	return http.StatusOK, nil
}

func (g Github) handleSignature(r *http.Request, body []byte, secret string) error {
	signature := r.Header.Get("X-Hub-Signature")
	if signature != "" {
		if secret == "" {
			return fmt.Errorf("empty webhook secret")
		} else {
			mac := hmac.New(sha1.New, []byte(secret))
			mac.Write(body)
			expectedMac := hex.EncodeToString(mac.Sum(nil))

			if signature[5:] != expectedMac {
				return fmt.Errorf("inavlid signature")
			}
		}
	}

	return nil
}

func (g Github) handlePush(body []byte, hc *HookConf) error {
	var push ghPush

	err := json.Unmarshal(body, &push)
	if err != nil {
		return err
	}

	refName := plumbing.ReferenceName(push.Ref)
	if refName.IsBranch() {
		if refName != hc.RefName {
			return fmt.Errorf("event: push to branch %s", refName)
		}
	} else {
		return fmt.Errorf("refName is not a branch: %s", refName)
	}
	return nil
}

func (g Github) handleRelease(body []byte, hc *HookConf) error {
	var release ghRelease

	err := json.Unmarshal(body, &release)
	if err != nil {
		return err
	}
	if release.Release.TagName == "" {
		return fmt.Errorf("invalid (empty) tag name")
	}

	return nil
}
