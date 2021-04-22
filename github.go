package caddy_webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Github struct {
}

type pushBody struct {
	Ref string `json:"ref"`
}

func (Github) Handle(req *http.Request, hc *HookConf) (int, error) {
	if err := ValidateRequest(req); err != nil {
		return http.StatusBadRequest, err
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return http.StatusRequestTimeout, err
	}

	signature := req.Header.Get("X-Hub-Signature")
	if signature != "" {
		if hc.Secret == "" {
			return http.StatusBadRequest, fmt.Errorf("empty webhook secret")
		}

		mac := hmac.New(sha1.New, []byte(hc.Secret))
		mac.Write(body)
		expectedMac := hex.EncodeToString(mac.Sum(nil))

		if signature[5:] != expectedMac {
			return http.StatusBadRequest, fmt.Errorf("inavlid signature")
		}
	}

	event := req.Header.Get("X-Github-Event")
	if event == "" {
		return http.StatusBadRequest, fmt.Errorf("header 'X-Github-Event' missing")
	}

	switch event {
	case "ping":
	case "push":
		var rBody pushBody

		err = json.Unmarshal(body, &rBody)
		if err != nil {
			return http.StatusBadRequest, err
		}
	default:
		return http.StatusBadRequest, fmt.Errorf("cannot handle %q event", event)
	}

	return http.StatusOK, nil
}
