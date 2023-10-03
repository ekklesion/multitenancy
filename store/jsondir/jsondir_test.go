package jsondir_test

import (
	"context"
	"github.com/ekklesion/multitenancy/store/jsondir"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSource_GetSite(t *testing.T) {
	uri, err := url.Parse("json+dir:testdata?watch=false")
	if err != nil {
		t.Fatalf(err.Error())
	}

	source, err := jsondir.CreateJsonSource(uri)
	if err != nil {
		t.Fatalf(err.Error())
	}

	defer func(c io.Closer) {
		_ = c.Close()
	}(source)

	ctx := context.Background()

	err = source.Initialize(ctx)
	if err != nil {
		t.Fatalf(err.Error())
	}

	req := httptest.NewRequest("GET", "/", http.NoBody)
	req.Host = "one.cloud.localhost"

	site, err := source.GetSite(ctx, req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if site.Id != "one.cloud.localhost" {
		t.Errorf("invalid site")
	}
}
