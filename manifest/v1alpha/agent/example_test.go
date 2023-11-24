package agent_test

import (
	"context"
	"log"

	"github.com/nobl9/nobl9-go/internal/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha/agent"
)

func ExampleAgent() {
	url := "https://prometheus-service.monitoring:8080"
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
				URL:    &url,
				Region: "",
			},
		},
	)
	// Verify the object:
	if err := myAgent.Validate(); err != nil {
		log.Fatalf("project validation failed, err: %v", err)
	}
	// Apply the object:
	client := examples.GetOfflineEchoClient()
	if err := client.ApplyObjects(context.Background(), []manifest.Object{myAgent}); err != nil {
		log.Fatalf("failed to apply project, err: %v", err)
	}
	// Output:
	// apiVersion: n9/v1alpha
	// kind: Agent
	// metadata:
	//   name: my-agent
	//   displayName: My Agent
	//   project: default
	// spec:
	//   description: Example project
	//   prometheus:
	//     url: https://prometheus-service.monitoring:8080
}
