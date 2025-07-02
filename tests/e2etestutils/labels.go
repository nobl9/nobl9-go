package e2etestutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// removeLabelByKey deletes a label by key.
func removeLabelByKey(labelKey string) {
	labelID, err := getLabelIDByKey(labelKey)
	if err != nil {
		printErrorf("failed to get label ID by name: %v", err)
		return
	}

	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode([]string{labelID}); err != nil {
		printErrorf("failed to encode cleanup labels payload: %v", err)
		return
	}
	req, err := sdkClient.CreateRequest(
		context.Background(),
		http.MethodPost,
		"labels/delete",
		nil,
		nil,
		&buf,
	)
	if err != nil {
		printErrorf("failed to create cleanup labels request: %v", err)
		return
	}
	resp, err := sdkClient.Do(req)
	if err != nil {
		printErrorf("failed to send cleanup labels request: %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		rawErr, _ := io.ReadAll(resp.Body)
		printErrorf("failed to cleanup labels, code: %d, body: %s", resp.StatusCode, string(rawErr))
		return
	}
}

func getLabelIDByKey(key string) (string, error) {
	req, err := sdkClient.CreateRequest(
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
	resp, err := sdkClient.Do(req)
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
		if label.Key == key {
			return label.ID, nil
		}
	}
	return "", fmt.Errorf("label '%s' not found", key)
}

func printErrorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}
