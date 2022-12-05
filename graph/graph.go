package graphqlclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gempages/go-shopify-graphql/graphql"
)

const (
	shopifyBaseDomain                  = "myshopify.com"
	shopifyAccessTokenHeader           = "X-Shopify-Access-Token"
	shopifyStoreFrontAccessTokenHeader = "X-Shopify-Storefront-Access-Token"
)

var (
	apiProtocol   = "https"
	apiPathPrefix = "admin/api"
	apiEndpoint   = "graphql.json"
)

// Option is used to configure options
type Option func(t *transport)

// WithContext optionally sets the API version if the passed string is valid
func WithContext(ctx context.Context) Option {
	return func(t *transport) {
		t.ctx = ctx
	}
}

// WithVersion optionally sets the API version if the passed string is valid
func WithVersion(apiVersion string) Option {
	return func(t *transport) {
		if apiVersion != "" {
			apiPathPrefix = fmt.Sprintf("admin/api/%s", apiVersion)
		} else {
			apiPathPrefix = "admin/api"
		}
	}
}

func WithStorefrontVersion(apiVersion string) Option {
	return func(t *transport) {
		if apiVersion != "" {
			apiPathPrefix = fmt.Sprintf("api/%s", apiVersion)
		} else {
			apiPathPrefix = "api"
		}
	}
}

// WithToken optionally sets oauth token
func WithToken(token string) Option {
	return func(t *transport) {
		t.accessToken = token
	}
}

// WithStoreFrontToken optionally sets storefront token

func WithStorefrontToken(token string) Option {
	return func(t *transport) {
		t.storefrontToken = token
	}
}

// WithPrivateAppAuth optionally sets private app credentials
func WithPrivateAppAuth(apiKey string, token string) Option {
	return func(t *transport) {
		t.apiKey = apiKey
		t.accessToken = token
	}
}

type transport struct {
	ctx             context.Context
	accessToken     string
	storefrontToken string
	apiKey          string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	isAccessTokenSet := t.accessToken != ""
	isStorefrontTokenSet := t.storefrontToken != ""
	areBasicAuthCredentialsSet := t.apiKey != "" && isAccessTokenSet

	if isAccessTokenSet {
		req.Header.Set(shopifyAccessTokenHeader, t.accessToken)
	} else if areBasicAuthCredentialsSet {
		req.SetBasicAuth(t.apiKey, t.accessToken)
	} else if isStorefrontTokenSet {
		req.Header.Set(shopifyStoreFrontAccessTokenHeader, t.storefrontToken)
	}

	return http.DefaultTransport.RoundTrip(req)
}

// NewClient creates a new client (in fact, just a simple wrapper for a graphql.Client)
func NewClient(shopName string, opts ...Option) *graphql.Client {
	trans := &transport{}

	for _, opt := range opts {
		opt(trans)
	}

	httpClient := &http.Client{Transport: trans}

	url := buildAPIEndpoint(shopName)

	graphClient := graphql.NewClient(url, httpClient)
	if trans.ctx != nil {
		graphClient.SetContext(trans.ctx)
	}

	return graphClient
}

func buildAPIEndpoint(shopName string) string {
	return fmt.Sprintf("%s://%s/%s/%s", apiProtocol, shopName, apiPathPrefix, apiEndpoint)
}
