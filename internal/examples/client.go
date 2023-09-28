package examples

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/goccy/go-yaml"

	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	"github.com/nobl9/nobl9-go/sdk"
)

// GetOfflineEchoClient creates an offline (local mock server) sdk.Client without auth (DisableOkta option).
// It is used exclusively for running code examples without internet connection or valid Nobl9 credentials.
// The body received by the server is decoded to JSON, converted to YAML and printed to stdout.
func GetOfflineEchoClient() *sdk.Client {
	// Offline server:
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p []v1alpha.Project
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			panic(err)
		}
		data, err := yaml.Marshal(p[0])
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	}))
	// Create sdk.Client:
	u, _ := url.Parse(srv.URL)
	config := &sdk.Config{DisableOkta: true, URL: u}
	client, err := sdk.NewClient(config)
	if err != nil {
		panic(err)
	}
	return client
}
