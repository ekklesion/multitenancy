package store

import (
	"context"
	"errors"
	"io"
	"net/http"
)

// Tenant represents the information about a tenant
type Tenant struct {
	Id     string            `json:"id,omitempty"`
	Domain string            `json:"domain,omitempty"`
	Status string            `json:"status,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

var ErrTenantNotFound = errors.New("tenant not found in source")
var ErrNotInitialized = errors.New("source not initialized")
var ErrAlreadyInitialized = errors.New("source already initialized")

// Source represent a source of store.Tenant
//
// The source CAN be implemented with any underlying storage.
type Source interface {
	io.Closer
	// GetSite returns the site information for the given request.
	//
	// This method should be very performant, ideally a map to memory lookup. The heavy lifting about keeping the
	// memory in sync should be done by the Initialize method.
	//
	// ErrTenantNotFound MUST be returned if the site does not exist
	//
	// ErrNotInitialized SHOULD be returned if Initialize has not been called. However, you can defer initialization
	// until a site is requested if you so desire.
	GetSite(ctx context.Context, r *http.Request) (*Tenant, error)

	// Initialize initializes this source.
	//
	// This method MUST be called in Caddy's provisioning phase
	// ErrAlreadyInitialized MUST be returned if this method is called twice
	Initialize(ctx context.Context) error
}
