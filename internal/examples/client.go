package examples

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
)

// GetOfflineEchoClient creates an offline (local mock server) [sdk.Client] without auth (DisableOkta option).
// It is used exclusively for running code examples without internet connection or valid Nobl9 credentials.
// The body received by the server is decoded to JSON, converted to YAML and printed to stdout.
func GetOfflineEchoClient() *sdk.Client {
	// Offline server:
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		objects, err := sdk.ReadObjectsFromSources(r.Context(), sdk.NewObjectSourceReader(r.Body, ""))
		if err != nil {
			panic(err)
		}
		if err = sdk.PrintObject(objects[0], os.Stdout, manifest.ObjectFormatYAML); err != nil {
			panic(err)
		}
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

// GetStaticClient creates an offline (local mock server) [sdk.Client] without auth (DisableOkta option).
// It is used exclusively for running code examples without internet connection or valid Nobl9 credentials.
// The response provided when initializing the client is returned by the server as JSON.
func GetStaticClient(response any) *sdk.Client {
	// Offline server:
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		enc.SetIndent("", " ")
		if err := enc.Encode(response); err != nil {
			panic(err)
		}
		w.WriteHeader(200)
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
