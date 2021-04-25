package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"io/ioutil"
	"net/http"
)

type Gogs struct {
}

type gogsPush struct {
	Ref string `json:"ref"`
}

func (g Gogs) Handle(r *http.Request, hc *HookConf) (int, error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	err = g.handleSignature(r, body, hc.Secret)
	if err != nil {
		return http.StatusBadRequest, err
	}

	event := r.Header.Get("X-Gogs-Event")
	if event == "" {
		return http.StatusBadRequest, fmt.Errorf("header 'X-Gogs-Event' missing")
	}

	switch event {
	case "push":
		err = g.handlePush(body, hc)
		if err != nil {
			return http.StatusBadRequest, err
		}
	default:
		return http.StatusBadRequest, fmt.Errorf("cannot handle %q event", event)
	}

	return http.StatusOK, nil
}

func (g Gogs) handleSignature(r *http.Request, body []byte, secret string) error {
	signature := r.Header.Get("X-Gogs-Signature")
	if signature != "" {
		if secret == "" {
			return fmt.Errorf("empty webhook secret")
		} else {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			expectedMac := hex.EncodeToString(mac.Sum(nil))

			if signature != expectedMac {
				return fmt.Errorf("inavlid signature")
			}
		}
	}

	return nil
}

func (g Gogs) handlePush(body []byte, hc *HookConf) error {
	var push gogsPush

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
