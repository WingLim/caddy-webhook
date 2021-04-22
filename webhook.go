package caddy_webhook

import (
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
	Handle(*http.Request, *WebHook) (int, error)
}

type WebHook struct {
	Repository string `json:"repo,omitempty"`
	Path       string `json:"path,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Type       string `json:"type,omitempty"`
	Secret     string `json:"secret,omitempty"`
	Depth      string `json:"depth,omitempty"`

	log *zap.Logger
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

func (w *WebHook) ServeHTTP(writer http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	return nil
}
