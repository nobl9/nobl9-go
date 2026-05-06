package sdk

import (
	"crypto/tls"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCustomCATransport(t *testing.T) {
	t.Parallel()
	t.Run("empty path returns nil transport", func(t *testing.T) {
		t.Parallel()
		rt, err := newCustomCATransport("")
		require.NoError(t, err)
		assert.Nil(t, rt)
	})

	t.Run("missing file returns wrapped error", func(t *testing.T) {
		t.Parallel()
		rt, err := newCustomCATransport(filepath.Join(t.TempDir(), "does-not-exist.pem"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read CA bundle")
		assert.Nil(t, rt)
	})

	t.Run("garbage file returns parse error", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "garbage.pem")
		require.NoError(t, os.WriteFile(path, []byte("not a pem file"), 0o600))
		rt, err := newCustomCATransport(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse any PEM certificates")
		assert.Nil(t, rt)
	})

	t.Run("trusts certificate from configured bundle", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(srv.Close)

		path := filepath.Join(t.TempDir(), "bundle.pem")
		require.NoError(t, os.WriteFile(path, leafCertPEM(t, srv.TLS), 0o600))

		rt, err := newCustomCATransport(path)
		require.NoError(t, err)
		require.NotNil(t, rt)

		client := &http.Client{Transport: rt}
		resp, err := client.Get(srv.URL)
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}

// TestNewClient_CACertFile verifies that [NewClient] surfaces an error from the
// CA bundle loader rather than silently falling back to the default trust store
// when [Config.CACertFile] is set but cannot be parsed.
func TestNewClient_CACertFile(t *testing.T) {
	t.Parallel()

	t.Run("garbage CA bundle aborts construction", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "garbage.pem")
		require.NoError(t, os.WriteFile(path, []byte("not a pem file"), 0o600))

		conf, err := ReadConfig(
			ConfigOptionNoConfigFile(),
			ConfigOptionWithCredentials("id", "secret"),
		)
		require.NoError(t, err)
		conf.CACertFile = path

		client, err := NewClient(conf)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "failed to parse any PEM certificates")
	})

	t.Run("empty CA bundle leaves construction unchanged", func(t *testing.T) {
		t.Parallel()
		conf, err := ReadConfig(
			ConfigOptionNoConfigFile(),
			ConfigOptionWithCredentials("id", "secret"),
		)
		require.NoError(t, err)
		assert.Empty(t, conf.CACertFile)

		client, err := NewClient(conf)
		require.NoError(t, err)
		require.NotNil(t, client)
	})
}

// leafCertPEM extracts the leaf certificate from a [tls.Config] (as exposed by
// [httptest.Server.TLS]) and returns it as a PEM-encoded byte slice.
func leafCertPEM(t *testing.T, cfg *tls.Config) []byte {
	t.Helper()
	require.NotNil(t, cfg)
	require.NotEmpty(t, cfg.Certificates)
	leaf := cfg.Certificates[0].Certificate[0]
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf})
}
