package caddy_webhook

import (
	"context"
	"fmt"
	"github.com/WingLim/caddy-webhook/webhooks"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Interface guards.
var (
	_ caddy.Module                = (*WebHook)(nil)
	_ caddy.Provisioner           = (*WebHook)(nil)
	_ caddy.Validator             = (*WebHook)(nil)
	_ caddyhttp.MiddlewareHandler = (*WebHook)(nil)
)

func init() {
	caddy.RegisterModule(new(WebHook))
	httpcaddyfile.RegisterHandlerDirective("webhook", parseHandlerCaddyfile)
}

// WebHook is the module configuration.
type WebHook struct {
	Repository  string   `json:"repo,omitempty"`
	Path        string   `json:"path,omitempty"`
	Branch      string   `json:"branch,omitempty"`
	Type        string   `json:"type,omitempty"`
	Secret      string   `json:"secret,omitempty"`
	Depth       string   `json:"depth,omitempty"`
	Submodule   bool     `json:"submodule,omitempty"`
	Command     []string `json:"command,omitempty"`
	Key         string   `json:"key,omitempty"`
	KeyPassword string   `json:"key_password,omitempty"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	Token       string   `json:"token,omitempty"`

	hook  webhooks.HookService
	auth  transport.AuthMethod
	cmd   *Cmd
	depth int
	repo  *Repo
	log   *zap.Logger
	ctx   context.Context
	setup bool
}

// CaddyModule returns the Caddy module information.
func (*WebHook) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.webhook",
		New: func() caddy.Module {
			return new(WebHook)
		},
	}
}

// Provision set's up webhook configuration.
func (w *WebHook) Provision(ctx caddy.Context) error {
	w.log = ctx.Logger(w)
	w.ctx = ctx.Context
	var err error

	if w.Path == "" {
		// If the path is empty for a repo, try to get the repo name from
		// the Repository. If successful set it to "./<repo-name>" else
		// set it to current working directory, i.e., "."
		name, err := getRepoNameFromURL(w.Repository)
		if err != nil {
			w.Path = "."
		} else {
			w.Path = name
		}
	}

	// Get the absolute path for logging
	w.Path, err = filepath.Abs(w.Path)
	if err != nil {
		return err
	}

	w.setHookType()

	// Convert depth from string to int
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

	if w.Command != nil {
		w.cmd = &Cmd{}
		w.cmd.AddCommand(w.Command, w.Path)
	}

	if w.Username != "" && w.Password != "" {
		w.auth = &githttp.BasicAuth{
			Username: w.Username,
			Password: w.Password,
		}
	}

	if w.Token != "" {
		w.auth = &githttp.BasicAuth{
			Username: "git", // This can be anything.
			Password: w.Token,
		}
	}

	if w.Key != "" {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", w.Key, w.KeyPassword)
		if err != nil {
			return err
		}
		w.auth = publicKeys
	}

	w.repo = NewRepo(w)

	if w.Submodule {
		w.repo.Submodule = git.DefaultSubmoduleRecursionDepth
	} else {
		w.repo.Submodule = git.NoRecurseSubmodules
	}
	return nil
}

// Validate ensures webhook's configuration is valid.
func (w *WebHook) Validate() error {
	if w.Repository == "" {
		return fmt.Errorf("cannot create repository with empty URL")
	}

	if w.Path == "" {
		return fmt.Errorf("cannot create repository in empty path")
	}

	if w.Key != "" && w.auth.Name() != ssh.PublicKeysName {
		return fmt.Errorf("wrong auth method with public key")
	}

	if w.Username != "" && w.Password != "" && w.auth.Name() != "http-basic-auth" {
		return fmt.Errorf("wrong auth method with username and password")
	}

	if w.Token != "" && w.auth.Name() != "http-basic-auth" {
		return fmt.Errorf("wrong auth method with token")
	}

	if !isEmptyOrGit(w.Path, w.log) {
		return fmt.Errorf("given path is neither empty nor git repository")
	}

	go func(webhook *WebHook) {
		if err := webhook.repo.Setup(webhook.ctx); err != nil {
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

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (w *WebHook) ServeHTTP(rw http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if !w.setup {
		return caddyhttp.Error(
			http.StatusNotFound,
			fmt.Errorf("page not found"),
		)
	}

	if err := ValidateRequest(r); err != nil {
		return err
	}

	hc := &webhooks.HookConf{
		Secret:  w.Secret,
		RefName: w.repo.refName,
	}

	code, err := w.hook.Handle(r, hc)
	if err != nil {
		rw.WriteHeader(code)
		w.log.Error(err.Error())
		return caddyhttp.Error(code, err)
	}

	go func(webhook *WebHook) {
		webhook.log.Info("updating repository", zap.String("path", webhook.Path))

		if err := webhook.repo.Update(webhook.ctx); err != nil {
			if err == git.NoErrAlreadyUpToDate {
				webhook.log.Info("already up-to-date", zap.String("path", webhook.Path))
			} else {
				webhook.log.Error(
					"cannot update repository",
					zap.Error(err),
					zap.String("path", webhook.Path),
				)
			}
			return
		}
	}(w)

	return nil
}

// setHookType set the type which hook service we will use.
func (w *WebHook) setHookType() {
	switch w.Type {
	case "gitee":
		w.hook = webhooks.Gitee{}
	case "gitlab":
		w.hook = webhooks.Gitlab{}
	case "bitbucket":
		w.hook = webhooks.Bitbucket{}
	case "gogs":
		w.hook = webhooks.Gogs{}
	default:
		w.hook = webhooks.Github{}
	}
}

// ValidateRequest validates webhook request, the webhook request
// should be POST.
func ValidateRequest(r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("only %s method accepted; got %s", http.MethodPost, r.Method)
	}

	return nil
}

// getRepoNameFromURL extracts the repo name from the HTTP URL of the repo.
func getRepoNameFromURL(u string) (string, error) {
	var name string
	if strings.HasPrefix(u, "http") {
		// Get repo name from http or https link.
		// https://github.com/WingLim/caddy-webhook.git

		netUrl, err := url.ParseRequestURI(u)
		if err != nil {
			return "", err
		}

		pathSegments := strings.Split(netUrl.Path, "/")
		name = pathSegments[len(pathSegments)-1]
	} else if strings.HasPrefix(u, "git") {
		// Get repo name from ssh link.
		// git@github.com:WingLim/caddy-webhook.git
		pathSegments := strings.Split(u, "/")
		name = pathSegments[len(pathSegments)-1]
	} else {
		return "", fmt.Errorf("unsupported protocol")
	}

	return strings.TrimSuffix(name, ".git"), nil
}

// isEmptyOrGit will check the path. If the path is neither empty nor a git
// directory, return error.
func isEmptyOrGit(path string, log *zap.Logger) bool {
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
		defer func(f *os.File) {
			if err := f.Close(); err != nil {
				log.Error(err.Error())
			}
		}(f)

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
