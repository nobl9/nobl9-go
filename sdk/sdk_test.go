package sdk

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetObject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
	}))
	defer srv.Close()

	client, err := DefaultClient("client-id", "super-secret", srv.URL, "123", "sloctl")
	require.NoError(t, err)

	objects, err := client.GetObject(
		context.Background(),
		"my-project",
		ObjectService,
		"",
		map[string][]string{"team": {"green", "purple"}},
		"service1", "service2",
	)
	require.NoError(t, err)
	require.NotEmpty(t, objects)
}
