package sdk

import (
	"context"
	"embed"
	_ "embed"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed test_data/config
var configTestData embed.FS

const configTestDataPath = "test_data/config"

func TestReadConfig_FromConfigFile(t *testing.T) {
	setupConfigTestData(t)

	ReadConfig(context.Background())
	require.NoError(t, err)
}

func setupConfigTestData(t *testing.T) (tempDir string) {
	t.Helper()
	tempDir = os.TempDir()
	dirEntries, err := configTestData.ReadDir(configTestDataPath)
	require.NoError(t, err)
	for _, entry := range dirEntries {
		copyEmbeddedFileToTempFile(t, tempDir, entry.Name())
	}
	return tempDir
}

func copyEmbeddedFileToTempFile(t *testing.T, tmpDir string, filename string) {
	embeddedFile, err := configTestData.Open(filepath.Join(configTestDataPath, filename))
	require.NoError(t, err)
	defer func() { _ = embeddedFile.Close() }()

	tmpFile, err := os.Create(filepath.Join(tmpDir, filename))
	require.NoError(t, err)
	defer func() { _ = tmpFile.Close() }()

	_, err = io.Copy(tmpFile, embeddedFile)
	require.NoError(t, err)
}
