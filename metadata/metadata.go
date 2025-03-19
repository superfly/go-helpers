// Fly.io machines are provided with a Unix socket that proxies traffic to the
// machines API, injecting an authentication token that is authorized to modify
// machine metadata. This package provides helpers for accessing this
// functionality.
package metadata

import (
	"context"
	"net"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

const SocketPath = "/.fly/api"

var (
	dialer net.Dialer
)

func Transport() *http.Transport {
	t := cleanhttp.DefaultTransport()
	t.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, "unix", SocketPath)
	}

	return t
}

func Client() *http.Client {
	return &http.Client{Transport: Transport()}
}
