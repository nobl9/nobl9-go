package agent_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
	v2 "github.com/nobl9/nobl9-go/sdk/endpoints/objects/v2"
)

func ExampleAgent() {
	// Create the object:
	myAgent := agent.New(
		agent.Metadata{
			Name:        "my-agent",
			DisplayName: "My Agent",
			Project:     "default",
		},
		agent.Spec{
			Description: "Example Agent",
			Prometheus: &agent.PrometheusConfig{
				URL: "https://prometheus-service.monitoring:8080",
			},
		},
	)
	// Verify the object:
	if err := myAgent.Validate(); err != nil {
		log.Fatalf("agent validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.Objects().V2().Apply(
		context.Background(),
		v2.ApplyRequest{Objects: []manifest.Object{myAgent}},
	); err != nil {
		log.Fatalf("failed to apply agent, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Agent
	// metadata:
	//   name: my-agent
	//   displayName: My Agent
	//   project: default
	// spec:
	//   description: Example Agent
	//   prometheus:
	//     url: https://prometheus-service.monitoring:8080
}
