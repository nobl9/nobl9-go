package sdk

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/stretchr/testify/assert"

	"github.com/nobl9/nobl9-go/manifest"
	objectsV1 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v1"
)

var Like = matchers.Like
var Regex = matchers.Regex

type Map = matchers.MapMatcher
type S = matchers.S

func TestConsumerV4(t *testing.T) {
	err := log.SetLogLevel("DEBUG")
	assert.NoError(t, err)

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: "nobl9-client",
		Provider: "nobl9-api",
		Host:     "127.0.0.1",
		TLS:      false,
		PactDir:  "./../pacts/",
	})
	assert.NoError(t, err)

	name := "my-service"

	// Set up our expected interactions.
	err = mockProvider.
		AddInteraction().
		Given("state 1").
		GivenWithParameter(models.ProviderState{
			Name: "Service exists",
			Parameters: map[string]interface{}{
				"name": name,
			},
		}).
		UponReceiving("A request to get a service").
		WithRequest("GET", "/api/internal/get/service", func(b *consumer.V4RequestBuilder) {
			b.Query("name", Regex(name, "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"))

		}).
		WillRespondWith(200, func(b *consumer.V4ResponseBuilder) {
			b.Header("Content-Type", S("application/json")).
				JSONBody([]Map{
					{
						"apiVersion": Like(manifest.VersionV1alpha),
						"kind":       Like(manifest.KindService),
						"metadata": matchers.StructMatcher{
							"name":        Like(name),
							"displayName": Like("My service"),
							"project":     Like("My project"),
						},
						"spec": matchers.StructMatcher{
							"description": Like("My service description"),
						},
						"organization": Like("nobl9"),
					},
				})
		}).
		ExecuteTest(t, test)

	assert.NoError(t, err)
}

var test = func(config consumer.MockServerConfig) error {
	client, err := NewClient(&Config{
		ClientID:     "client-id",
		ClientSecret: "secret",
		AccessToken:  "token",
		DisableOkta:  true,
		Organization: "nobl9",
		URL: &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", "localhost", config.Port),
			Path:   "/api/internal",
		},
	})
	if err != nil {
		return err
	}

	services, err := client.Objects().V1().GetV1alphaServices(
		context.Background(),
		objectsV1.GetServicesRequest{
			Project: "My project",
			Names:   []string{"my-service"},
		},
	)

	if len(services) < 1 {
		return fmt.Errorf("no services")
	}

	return nil
}
