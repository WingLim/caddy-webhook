package caddy_webhook

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"net/http"
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

	Hook HookService
	log  *zap.Logger
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
	return nil
}

func (w *WebHook) Validate() error {
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
