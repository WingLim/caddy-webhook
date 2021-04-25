package webhooks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Bitbucket struct {
}

type bbPush struct {
	Push struct {
		Changes []struct {
			New struct {
				Type string `json:"type,omitempty"`
				Name string `json:"name,omitempty"`
			} `json:"new,omitempty"`
		} `json:"changes,omitempty"`
	} `json:"push,omitempty"`
}

func (b Bitbucket) Handle(r *http.Request, hc *HookConf) (int, error) {
	if !b.verifyBitbucketIP(r.RemoteAddr) {
		return http.StatusForbidden, fmt.Errorf("the request doesn't come from a valid IP")
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, err
	}

	event := r.Header.Get("X-Event-Key")
	if event == "" {
		return http.StatusBadRequest, fmt.Errorf("header 'X-Event-Key' missing")
	}

	switch event {
	case "repo:push":
		err = b.handlePush(body, hc)
		if err != nil {
			return http.StatusBadRequest, err
		}
	default:
		return http.StatusBadRequest, fmt.Errorf("cannot handle %q event", event)
	}
	return http.StatusOK, nil
}

func (b Bitbucket) handlePush(body []byte, hc *HookConf) error {
	var push bbPush

	err := json.Unmarshal(body, &push)
	if err != nil {
		return err
	}

	if len(push.Push.Changes) == 0 {
		return fmt.Errorf("the push was incomplete, missing change list")
	}

	change := push.Push.Changes[0]
	if len(change.New.Name) == 0 {
		return fmt.Errorf("the push didn't contain a valid branch name")
	}

	refType := change.New.Type
	if refType == "" {
		return fmt.Errorf("the push didn't cotain type")
	}

	refName := change.New.Name
	if refType == "branch" {
		if refName != hc.RefName.Short() {
			return fmt.Errorf("event: push to branch %s", refName)
		}
	} else {
		return fmt.Errorf("refName is not a branch: %s", refName)
	}

	return nil
}

func hostOnly(remoteAddr string) string {
	host, _, _ := net.SplitHostPort(remoteAddr)
	if host == "" {
		return remoteAddr
	}
	return host
}

func (b Bitbucket) verifyBitbucketIP(remoteAddr string) bool {
	ipAddress := net.ParseIP(hostOnly(remoteAddr))

	if err := updateBitBucketIPs(); err != nil && len(atlassianIPs.Items) == 0 {
		return false
	}

	atlassianIPsMu.Lock()
	ipItems := atlassianIPs.Items
	atlassianIPsMu.Unlock()

	for _, item := range ipItems {
		if !strings.Contains(item.CIDR, "") {
			ip := net.ParseIP(item.CIDR)
			if ip.Equal(ipAddress) {
				return true
			}
			continue
		}

		_, cidrnet, err := net.ParseCIDR(item.CIDR)
		if err != nil {
			continue
		}

		if cidrnet.Contains(ipAddress) {
			return true
		}
	}
	return false
}

type atlassianIPResponse struct {
	CreationDate string             `json:"creationDate"`
	SyncToken    int                `json:"syncToken"`
	Items        []atlassianIPRange `json:"items"`

	lastUpdated time.Time
}

type atlassianIPRange struct {
	Network string `json:"network"`
	MaskLen int    `json:"mask_len"`
	CIDR    string `json:"cidr"`
	Mask    string `json:"mask"`
}

var (
	atlassianIPs   atlassianIPResponse
	atlassianIPsMu sync.Mutex
)

func updateBitBucketIPs() error {
	atlassianIPsMu.Lock()
	defer atlassianIPsMu.Unlock()

	if atlassianIPs.lastUpdated.IsZero() || time.Since(atlassianIPs.lastUpdated) > 24*time.Hour {
		resp, err := http.Get("https://ip-ranges.atlassian.com/")
		if err != nil {
			return fmt.Errorf("fail to request recent IPs for bitbucket: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("fail to request recent IPs for bitbucket: HTTP %d", resp.StatusCode)
		}

		var newIPs atlassianIPResponse
		err = json.NewDecoder(resp.Body).Decode(&newIPs)
		if err != nil {
			return fmt.Errorf("fail to decode recent IPs for bitbucket: %v", err)
		}

		newIPs.lastUpdated = time.Now()
		atlassianIPs = newIPs
	}
	return nil
}
