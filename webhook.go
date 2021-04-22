package caddy_webhook

import (
	"context"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	_ caddy.Module                = (*WebHook)(nil)
	_ caddy.Provisioner           = (*WebHook)(nil)
	_ caddy.Validator             = (*WebHook)(nil)
	_ caddyhttp.MiddlewareHandler = (*WebHook)(nil)
)

type HookService interface {
	Handle(*http.Request, *HookConf) (int, error)
}

type WebHook struct {
	Repository string `json:"repo,omitempty"`
	Path       string `json:"path,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Type       string `json:"type,omitempty"`
	Secret     string `json:"secret,omitempty"`
	Depth      string `json:"depth,omitempty"`

	Hook  HookService
	repo  *Repo
	log   *zap.Logger
	ctx   context.Context
	setup bool
}

type HookConf struct {
	Secret string
}

func (*WebHook) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.webhook",
		New: func() caddy.Module {
			return new(WebHook)
		},
	}
}

func (w *WebHook) Provision(ctx caddy.Context) error {
	w.log = ctx.Logger(w)
	w.ctx = ctx.Context
	var err error
	if w.Path == "" {
		name, err := getRepoNameFromURL(w.Repository)
		if err != nil {
			w.Path = "."
		} else {
			w.Path = name
		}
	}
	w.Path, err = filepath.Abs(w.Path)
	if err != nil {
		return err
	}

	if w.Type == "" {
		w.Hook = Github{}
	}

	var depth int
	if w.Depth != "" {
		depth, err = strconv.Atoi(w.Depth)
		if err != nil {
			return err
		}
	} else {
		depth = 0
	}

	w.repo = &Repo{
		URL:    w.Repository,
		Path:   w.Path,
		Branch: w.Branch,
		Depth:  depth,
	}
	return nil
}

func (w *WebHook) Validate() error {
	if w.Repository == "" {
		return fmt.Errorf("cannot create repository with empty URL")
	}

	u, err := url.Parse(w.Repository)
	if err != nil {
		return fmt.Errorf("invalid url: %v", err)
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("url scheme '%s' not supported", u.Scheme)
	}

	if w.Path == "" {
		return fmt.Errorf("cannot create repository in empty path")
	}
	if err := isEmptyOrGit(w.Path); err != nil {
		return fmt.Errorf("given path is neither empty nor git repository")
	}

	go func(webhook *WebHook) {
		if err := webhook.repo.Setup(webhook.ctx, webhook.log); err != nil {
			webhook.log.Error(
				"repository not setup",
				zap.Error(err),
				zap.String("path", webhook.Path))
			return
		}
		webhook.setup = true
	}(w)

	return nil
}

func (w *WebHook) ServeHTTP(rw http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	hc := HookConf{Secret: w.Secret}

	code, err := w.Hook.Handle(r, &hc)
	if err != nil {
		rw.WriteHeader(code)
		return caddyhttp.Error(code, err)
	}

	return next.ServeHTTP(rw, r)
}

func ValidateRequest(r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("only %s method accepted; got %s", http.MethodPost, r.Method)
	}

	return nil
}

func getRepoNameFromURL(u string) (string, error) {
	netUrl, err := url.ParseRequestURI(u)
	if err != nil {
		return "", err
	}

	pathSegments := strings.Split(netUrl.Path, "/")
	name := pathSegments[len(pathSegments)-1]
	return strings.TrimSuffix(name, ".git"), nil
}

func isEmptyOrGit(root string) error {
	info, err := os.Stat(root)
	if err != nil && err != os.ErrNotExist {
		return err
	}
	if info != nil && !info.IsDir() {
		return fmt.Errorf("path is not a dir")
	}

	_, err = git.PlainOpen(root)
	if err != nil {
		return err
	}
	return nil
}
