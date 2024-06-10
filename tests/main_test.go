//go:build e2e_test

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/nobl9/nobl9-go/sdk"
)

var client *sdk.Client

func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	var err error
	if client, err = sdk.DefaultClient(); err != nil {
		printError("failed to create %T: %v", client, err)
		return 1
	}
	org, err := client.GetOrganization(context.Background())
	if err != nil {
		printError("failed to get test organization: %v", err)
		return 1
	}
	fmt.Printf("Running SDK end-to-end tests\nOrganization: %s\nEnvironment: %s\nClient ID: %s\n\n",
		org, client.Config.URL, client.Config.ClientID)
	defer cleanupLabels()
	return m.Run()
}

// cleanupLabels deletes all unique labels created during the test.
func cleanupLabels() {
	labelID, err := getLabelIDByName(uniqueTestIdentifierLabel.Key)
	if err != nil {
		printError("failed to get label ID by name: %v", err)
		return
	}

	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode([]string{labelID}); err != nil {
		printError("failed to encode cleanup labels payload: %v", err)
		return
	}
	req, err := client.CreateRequest(
		context.Background(),
		http.MethodPost,
		"labels/delete",
		nil,
		nil,
		&buf,
	)
	if err != nil {
		printError("failed to create cleanup labels request: %v", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		printError("failed to send cleanup labels request: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		rawErr, _ := io.ReadAll(resp.Body)
		printError("failed to cleanup labels, code: %d, body: %s", resp.StatusCode, string(rawErr))
		return
	}
}

func getLabelIDByName(name string) (string, error) {
	req, err := client.CreateRequest(
		context.Background(),
		http.MethodGet,
		"labels",
		nil,
		nil,
		nil,
	)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		rawErr, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to cleanup labels, code: %d, body: %s", resp.StatusCode, string(rawErr))
	}
	var labels []struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&labels); err != nil {
		return "", err
	}
	for _, label := range labels {
		if label.Key == name {
			return label.ID, nil
		}
	}
	return "", fmt.Errorf("label '%s' not found", name)
}

func printError(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}
