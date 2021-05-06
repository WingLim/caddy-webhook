package caddy_webhook

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func parseHandlerCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	wh := new(WebHook)
	err := wh.UnmarshlCaddyfile(h.Dispenser)
	if err != nil {
		return nil, err
	}
	return wh, nil
}

//	UnmarshCaddyfile configures the handler directive from Caddyfile.
//	Syntax:
//
//		webhook [<url> <path>] {
//			repo		<text>
//			path 		<text>
//			branch 		<text>
//			depth		<int>
//			type 		<text>
//			secret		<text>
//			command		<test>...
//			submodule
//		}
func (w *WebHook) UnmarshlCaddyfile(d *caddyfile.Dispenser) error {
	if d.NextArg() && d.NextArg() {
		w.Repository = d.Val()
	}

	if d.NextArg() {
		w.Path = d.Val()
	}

	for d.NextBlock(0) {
		switch d.Val() {
		case "repo":
			if w.Repository != "" {
				return d.Err("url specified twice")
			}
			if !d.Args(&w.Repository) {
				return d.ArgErr()
			}
		case "path":
			if w.Path != "" {
				return d.Err("path specified twice")
			}
			if !d.Args(&w.Path) {
				return d.ArgErr()
			}
		case "branch":
			if !d.Args(&w.Branch) {
				return d.ArgErr()
			}
		case "depth":
			if !d.Args(&w.Depth) {
				return d.ArgErr()
			}
		case "type":
			if !d.Args(&w.Type) {
				return d.ArgErr()
			}
		case "secret":
			if !d.Args(&w.Secret) {
				return d.ArgErr()
			}
		case "submodule":
			w.Submodule = true
		case "command":
			w.Command = d.RemainingArgs()
		case "key":
			if !d.Args(&w.Key) {
				return d.ArgErr()
			}
		case "key_password":
			if !d.Args(&w.KeyPassword) {
				return d.ArgErr()
			}

		}
	}

	return nil
}
