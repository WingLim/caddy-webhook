package caddy_webhook

import (
	"fmt"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/go-git/go-git/v5"
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

func TestIsEmptyOrGit(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"test/not_exist", true},
		{"test/empty", true},
		{"test/not_empty", false},
		{"test/file", false},
		{"test/dir_git", true},
	}

	err := os.Mkdir("test", 0666)
	defer func() {
		err := os.RemoveAll("test")
		assert.Nil(t, err)
	}()
	assert.Nil(t, err)

	err = os.Mkdir("test/empty", 0666)
	assert.Nil(t, err)

	err = os.Mkdir("test/not_empty", 0666)
	assert.Nil(t, err)
	file1, err := os.Create("test/not_empty/test.txt")
	defer func(file1 *os.File) {
		err := file1.Close()
		assert.Nil(t, err)
	}(file1)
	assert.Nil(t, err)

	file2, err := os.Create("test/file")
	defer func(file2 *os.File) {
		err := file2.Close()
		assert.Nil(t, err)
	}(file2)
	assert.Nil(t, err)

	err = os.Mkdir("test/dir", 0666)
	assert.Nil(t, err)

	_, err = git.PlainInit("test/dir_git", false)
	assert.Nil(t, err)

	for i, tc := range testCases {
		actual := isEmptyOrGit(tc.path, nil)

		assert.Equal(t, tc.expected, actual, fmt.Sprintf("case %d", i))
	}
}
