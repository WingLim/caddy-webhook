package webhooks

import (
	"github.com/go-git/go-git/v5/plumbing"
	"net/http"
)

type HookConf struct {
	Secret string

	RefName plumbing.ReferenceName
}

type HookService interface {
	Handle(*http.Request, *HookConf) (int, error)
}
