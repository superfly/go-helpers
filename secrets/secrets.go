// Provides helpers for fetching the latest version of app secrets.
package secrets

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/breml/rootcerts/embedded"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/superfly/go-helpers/metadata"
	"github.com/superfly/macaroon"
	"github.com/superfly/macaroon/bundle"
	"github.com/superfly/macaroon/flyio"
)

// Get fetches the secrets from the Fly.io secrets API. If no options are
// specified, all secrets are retrieved using a token that is automatically
// fetched from the machines API.
func Get(ctx context.Context, options ...Opt) (map[string]string, error) {
	const url = "https://petsem-public.fly.dev/petsem.v1.SecretService/GetSecrets"

	opts := &opts{}
	for _, o := range options {
		o(opts)
	}

	token, err := opts.getToken(ctx)
	if err != nil {
		return nil, err
	}

	appID, err := opts.getAppID(ctx, token)
	if err != nil {
		return nil, err
	}

	req := getSecretsRequest{
		AppId:    id{ID: appID},
		Selector: opts.getSelector(),
		Types:    opts.getTypes(),
		Versions: opts.versions,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	hReq.Header.Set("Authorization", token)
	hReq.Header.Set("Content-Type", "application/json")
	hReq.Header.Set("Accept", "application/json")

	hResp, err := opts.getClient().Do(hReq)
	if err != nil {
		return nil, err
	}
	defer hResp.Body.Close()

	if hResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", hResp.StatusCode)
	}

	var resp getSecretsResponse
	if err := json.NewDecoder(hResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	m := make(map[string]string, len(resp.Secrets))
	for _, s := range resp.Secrets {
		m[s.Label] = string(s.Value)
	}

	return m, nil
}

type opts struct {
	client   *http.Client
	token    string
	appID    int64
	versions map[string]uint64
	selector any
	types    []int32
}

func (o *opts) getClient() *http.Client {
	if o.client != nil {
		return o.client
	}

	t := cleanhttp.DefaultPooledTransport()
	t.TLSClientConfig = &tls.Config{RootCAs: x509.NewCertPool()}
	t.TLSClientConfig.RootCAs.AppendCertsFromPEM([]byte(embedded.MozillaCACertificatesPEM()))

	return &http.Client{Transport: t}
}

func (o *opts) getToken(ctx context.Context) (string, error) {
	if o.token != "" {
		return o.token, nil
	}

	return Token(ctx)
}

func (o *opts) getAppID(ctx context.Context, token string) (int64, error) {
	if o.appID != 0 {
		return o.appID, nil
	}

	bun, err := bundle.ParseBundle(flyio.LocationSecrets, token)
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	allCavs := bundle.Reduce(bun, func(cavs []macaroon.Caveat, m bundle.Macaroon) []macaroon.Caveat {
		return append(cavs, m.UnsafeCaveats().Caveats...)
	})

	appScope := flyio.AppScope(macaroon.NewCaveatSet(allCavs...))

	if len(appScope) != 1 {
		return 0, fmt.Errorf("failed to derive app id from token: %w", err)
	}

	return int64(appScope[0]), nil
}

func (o *opts) getSelector() any {
	if o.selector != nil {
		return o.selector
	}
	return secretsSelectorAll{All: true}
}

func (o *opts) getTypes() []int32 {
	const secretTypeAppSecret = 1

	if o.types != nil {
		return o.types
	}

	return []int32{secretTypeAppSecret}
}

type Opt func(*opts)

// WithClient sets the http.Client to use for requests. If not specified, a
// default is used.
func WithClient(client *http.Client) Opt {
	return func(o *opts) {
		o.client = client
	}
}

// WithToken sets the token to use for requests. If not specified, a new token
// is fetched.
func WithToken(token string) Opt {
	return func(o *opts) {
		o.token = token
	}
}

// WithAppID sets the app ID to use for requests. If not specified, the app ID
// is derived from the token.
func WithAppID(appID int64) Opt {
	return func(o *opts) {
		o.appID = appID
	}
}

// WithVersions sets the secret versions to request.
func WithVersions(versions map[string]uint64) Opt {
	return func(o *opts) {
		o.versions = versions
	}
}

// Labels specifies the secrets to request. If not specified, all secrets are
// requested.
func Labels(labels ...string) Opt {
	return func(o *opts) {
		o.selector = secretsSelectLabels{
			Labels: struct {
				Labels []string `json:"labels"`
			}{Labels: labels},
		}
	}
}

// Types specifies the secret types to request. If not specified, only app
// secrets are requested.
func Types(types ...int32) Opt {
	return func(o *opts) {
		o.types = types
	}
}

type secretsSelectorAll struct {
	All bool `json:"all"`
}

type secretsSelectLabels struct {
	Labels secretsSelectLabelsLabels `json:"labels"`
}

type secretsSelectLabelsLabels struct {
	Labels []string `json:"labels"`
}

type id struct {
	ID int64 `json:"id"`
}

type getSecretsRequest struct {
	AppId    id                `json:"app_id"`
	Selector any               `json:"selector"`
	Versions map[string]uint64 `json:"versions,omitempty"`
	Types    []int32           `json:"types"`
}

type getSecretsResponse struct {
	Secrets []struct {
		Label string `json:"label"`
		Value []byte `json:"value"`
	} `json:"secrets"`
}

// Token fetches a token authorized to read secrets for the app.
func Token(ctx context.Context) (string, error) {
	const url = "http://flaps/v1/tokens/kms"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := metadata.Client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}
