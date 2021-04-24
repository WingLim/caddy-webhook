package webhooks

import (
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"io/ioutil"
	"net/http"
)

type Gitlab struct {
}

type glPush struct {
	Ref string `json:"ref"`
}

func (g Gitlab) Handle(r *http.Request, hc *HookConf) (int, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	err = g.handleToken(r, hc.Secret)
	if err != nil {
		return http.StatusBadRequest, err
	}

	event := r.Header.Get("X-Gitlab-Event")
	if event == "" {
		return http.StatusBadRequest, fmt.Errorf("header 'X-Gitlab-Event' missing")
	}

	switch event {
	case "Push Hook":
		err = g.handlePush(body, hc)
	default:
		return http.StatusBadRequest, fmt.Errorf("cannot handle %q event", event)
	}

	return http.StatusOK, nil
}

func (g Gitlab) handleToken(r *http.Request, secret string) error {
	token := r.Header.Get("X-Gitlab-Token")
	if token != "" {
		if secret == "" {
			return fmt.Errorf("empty webhook secret")
		} else {
			if token != secret {
				return fmt.Errorf("inavlid token")
			}
		}
	}

	return nil
}

func (g Gitlab) handlePush(body []byte, hc *HookConf) error {
	var push glPush

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
