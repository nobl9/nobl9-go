// Package main provides a simple SDK client implementation.
// It is used to test the sdk.HeaderUserAgent defaults as
// debug.ReadBuildInfo requires a binary built from module to
// provide details such as the SDK package version.
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
)

func main() {
	config, err := sdk.ReadConfig(
		sdk.ConfigOptionWithCredentials("clientId", "clientSecret"),
		sdk.ConfigOptionNoConfigFile())
	config.DisableOkta = true
	if err != nil {
		panic(err)
	}
	rt := &mockRoundTripper{}
	client := sdk.NewClient(config)
	client.HTTP = &http.Client{Transport: rt}
	if err = client.ApplyObjects(context.Background(), []manifest.Object{}, false); err != nil {
		panic(err)
	}
	fmt.Print(rt.UserAgent)
}

type mockRoundTripper struct {
	UserAgent string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.UserAgent = req.Header.Get(sdk.HeaderUserAgent)
	return &http.Response{StatusCode: http.StatusOK}, nil
}
