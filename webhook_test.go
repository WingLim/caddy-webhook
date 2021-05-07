package caddy_webhook

import (
	"github.com/alecthomas/assert"
	"testing"
)

func TestGetRepoNameFromURL(t *testing.T) {
	testCases := []struct {
		link string
		name string
		err  bool
	}{
		{"https://github.com/WingLim/caddy-webhook.git", "caddy-webhook", false},
		{"git@github.com:WingLim/caddy-webhook.git", "caddy-webhook", false},
		{"ftp://balabala.git", "", true},
	}

	for _, tc := range testCases {
		name, err := getRepoNameFromURL(tc.link)

		if tc.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, tc.name, name)
		}
	}
}
