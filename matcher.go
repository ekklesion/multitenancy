package multitenancy

import (
	"errors"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/ekklesion/multitenancy/store"
	_ "github.com/ekklesion/multitenancy/store/jsondir"
	"go.uber.org/zap"
	"net/http"
	"net/url"
)

func init() {
	caddy.RegisterModule(&SiteInfoMatcher{})
}

const (
	VarSiteId     = "site_info.id"
	VarSiteStatus = "site_info.status"
)

// SiteInfoMatcher is a Caddy matcher that injects variables from a host
type SiteInfoMatcher struct {
	// The source of the store info. It can be anything the factory supports.
	SourceStr string `json:"source,omitempty"`
	source    store.Source
	logger    *zap.Logger
}

func (h SiteInfoMatcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.site_info",
		New: func() caddy.Module { return new(SiteInfoMatcher) },
	}
}

func (h *SiteInfoMatcher) Provision(ctx caddy.Context) error {
	h.logger = ctx.Logger(SiteInfoMatcher{})
	h.logger.Debug("Creating source", zap.String("source", h.SourceStr))

	uri, err := url.Parse(h.SourceStr)
	if err != nil {
		return err
	}

	h.source, err = store.CreateSource(uri)
	if err != nil {
		return err
	}

	h.logger.Debug("Initializing source", zap.String("source", h.SourceStr))
	err = h.source.Initialize(ctx)
	if err != nil {
		return err
	}

	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (h *SiteInfoMatcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.Args(&h.SourceStr) {
			return d.ArgErr()
		}
	}
	return nil
}

func (h *SiteInfoMatcher) Match(r *http.Request) bool {
	h.logger.Debug("Matching host", zap.String("host", r.Host))

	site, err := h.source.GetSite(r.Context(), r)
	if errors.Is(err, store.ErrTenantNotFound) {
		h.logger.Warn("Host was not found", zap.String("host", r.Host))
		return false
	}
	if err != nil {
		h.logger.Error("Match error", zap.Error(err))
		return false
	}

	h.logger.Info("Match found", zap.String("host", r.Host), zap.String("id", site.Id))

	rpl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	rpl.Set(VarSiteId, site.Id)
	rpl.Set(VarSiteStatus, site.Status)
	for key, value := range site.Params {
		rpl.Set(key, value)
	}

	return true
}

func (h *SiteInfoMatcher) Cleanup() error {
	return h.source.Close()
}
