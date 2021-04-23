package caddy_webhook

import (
	"context"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/go-git/go-git/v5"
	"go.uber.org/zap"
	"io"
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

func init() {
	caddy.RegisterModule(&WebHook{})
	httpcaddyfile.RegisterHandlerDirective("webhook", parseHandlerCaddyfile)
}

type HookService interface {
	Handle(*http.Request, *Repo) (int, error)
}

type WebHook struct {
	Repository string `json:"repo,omitempty"`
	Path       string `json:"path,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Type       string `json:"type,omitempty"`
	Secret     string `json:"secret,omitempty"`
	Depth      string `json:"depth,omitempty"`

	Hook  HookService
	depth int
	repo  *Repo
	log   *zap.Logger
	ctx   context.Context
	setup bool
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

	w.setHookType()

	var depth int
	if w.Depth != "" {
		depth, err = strconv.Atoi(w.Depth)
		if err != nil {
			return err
		}
	} else {
		depth = 0
	}
	w.depth = depth

	w.repo = NewRepo(w)
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
	if !isEmptyOrGit(w.Path) {
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
	if !w.setup {
		return caddyhttp.Error(
			http.StatusNotFound,
			fmt.Errorf("page not found"),
		)
	}

	code, err := w.Hook.Handle(r, w.repo)
	if err != nil {
		rw.WriteHeader(code)
		w.log.Error(err.Error())
		return caddyhttp.Error(code, err)
	}

	go func(webhook *WebHook) {
		webhook.log.Info("updating repository", zap.String("path", webhook.Path))

		if err := webhook.repo.Update(webhook.ctx); err != nil {
			webhook.log.Error(
				"cannot update repository",
				zap.Error(err),
				zap.String("path", webhook.Path),
			)
			return
		}
	}(w)

	return next.ServeHTTP(rw, r)
}

func (w *WebHook) setHookType() {
	switch w.Type {
	default:
		w.Hook = Github{}
	}
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

func isEmptyOrGit(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		return false
	}
	if info.IsDir() {
		f, err := os.Open(filepath.Clean(path))
		if err != nil {
			return false
		}
		defer f.Close()

		_, err = f.Readdirnames(1)
		if err == io.EOF {
			return true
		}
	}

	_, err = git.PlainOpen(path)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return false
		}
	}
	return true
}
