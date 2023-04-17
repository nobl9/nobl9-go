package definitions

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/nobl9-go/sdk"
)

//go:embed test_data
var testData embed.FS
var templates *template.Template

func TestMain(m *testing.M) {
	// Register templates.
	var err error
	templates, err = template.ParseFS(testData, "test_data/expected/*.tpl.json")
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestResolveSources(t *testing.T) {
	tmp := t.TempDir()
	for _, fn := range []string{"slo.yaml", "slo.yml", "slo.json", "slo.xml"} {
		_, err := os.Create(filepath.Join(tmp, fn))
		require.NoError(t, err)
	}

	rawSources := []RawSource{
		"http://insecure.com",
		"https://secure.com",
		tmp,
		filepath.Join(tmp, "**"),
		filepath.Join(tmp, "slo.json"),
	}

	expected := []*Source{
		{
			Type:  SourceTypeURL,
			Paths: []string{"http://insecure.com"},
			Raw:   "http://insecure.com",
		},
		{
			Type:  SourceTypeURL,
			Paths: []string{"https://secure.com"},
			Raw:   "https://secure.com",
		},
		{
			Type: SourceTypeDirectory,
			Paths: []string{
				filepath.Join(tmp, "slo.json"),
				filepath.Join(tmp, "slo.yaml"),
				filepath.Join(tmp, "slo.yml"),
			},
			Raw: tmp,
		},
		{
			Type: SourceTypeGlobPattern,
			Paths: []string{
				filepath.Join(tmp, "slo.json"),
				filepath.Join(tmp, "slo.yaml"),
				filepath.Join(tmp, "slo.yml"),
			},
			Raw: filepath.Join(tmp, "**"),
		},
		{
			Type:  SourceTypeFile,
			Paths: []string{filepath.Join(tmp, "slo.json")},
			Raw:   filepath.Join(tmp, "slo.json"),
		},
	}

	for _, raw := range rawSources {
		source, err := ResolveSource(raw)
		require.NoError(t, err)
		assert.Contains(t, expected, source)
	}

	sources, err := ResolveSources(rawSources...)
	require.NoError(t, err)
	assert.ElementsMatch(t, expected, sources)
}

func TestReadDefinitions_FromReader(t *testing.T) {
	t.Run("read definitions from reader", func(t *testing.T) {
		definitions, err := ReadSources(
			context.Background(),
			MetadataAnnotations{Organization: "my-org"},
			NewInputSource(readTestFile(t, "service_and_agent.yaml"), "stdin"))
		require.NoError(t, err)
		definitionsMatchExpected(t, definitions, expectedMeta{Name: "service_and_agent", ManifestSrc: "stdin"})
	})

	t.Run("read definitions from reader for empty source", func(t *testing.T) {
		definitions, err := ReadSources(
			context.Background(),
			MetadataAnnotations{Organization: "org"},
			NewInputSource(readTestFile(t, "service_and_agent.yaml"), "test"))
		require.NoError(t, err)
		definitionsMatchExpected(t,
			definitions,
			expectedMeta{Name: "service_and_agent", ManifestSrc: "test", Organization: "org"})
	})

	t.Run("report an error when io.Reader is nil", func(t *testing.T) {
		_, err := ReadSources(context.Background(), MetadataAnnotations{}, NewInputSource(nil, "nil"))
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrIoReaderIsNil)
	})
}

func TestReadDefinitions_FromURL(t *testing.T) {
	t.Run("successful definitions GET for http scheme", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := readTestFile(t, "annotations.yaml").WriteTo(w)
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		require.Regexp(t, "^http://", srv.URL)

		definitions, err := Read(context.Background(), MetadataAnnotations{Organization: "my-org"}, srv.URL)
		require.NoError(t, err)
		definitionsMatchExpected(t, definitions, expectedMeta{Name: "annotations", ManifestSrc: srv.URL})
	})

	t.Run("successful definitions GET for https scheme", func(t *testing.T) {
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := readTestFile(t, "annotations.yaml").WriteTo(w)
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		httpClientFactory = func(url string) *http.Client { return srv.Client() }
		require.Regexp(t, "^https://", srv.URL)

		definitions, err := Read(context.Background(), MetadataAnnotations{Organization: "org"}, srv.URL)
		require.NoError(t, err)
		definitionsMatchExpected(t,
			definitions,
			expectedMeta{Name: "annotations", ManifestSrc: srv.URL, Organization: "org"},
		)
	})

	t.Run("bad response status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "some error reason", http.StatusForbidden)
		}))
		httpClientFactory = func(url string) *http.Client { return srv.Client() }
		defer srv.Close()

		_, err := Read(context.Background(), MetadataAnnotations{Organization: "my-org"}, srv.URL)
		require.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("GET %s response: 403 some error reason", srv.URL))
	})

	t.Run("cancel request if context is cancelled", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		var err error
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = Read(ctx, MetadataAnnotations{Organization: "my-org"}, srv.URL)

		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestReadDefinitions_FromFS(t *testing.T) {
	ctx := context.Background()
	td := t.TempDir()
	sd := t.TempDir()
	wd, err := os.Getwd()
	require.NoError(t, err)

	// Helper functions:
	// - Get the temporary directory file path.
	tmpDir := func(path string) string { return filepath.Join(td, path) }
	// - Get the symlink directory file path.
	symlinkDir := func(path string) string { return filepath.Join(sd, path) }
	// - Get the working directory file path.
	workingDir := func(path string) string { return filepath.Join(wd, path) }
	// - Create temporary file.
	createFile := func(path string) {
		require.NoError(t, os.WriteFile(
			tmpDir(path),
			readTestFile(t, filepath.Base(path)).Bytes(),
			0o666))
	}
	// - Create temporary directory.
	createDir := func(path string) {
		require.NoError(t, os.Mkdir(
			tmpDir(path),
			0o777))
	}
	// - Create symlink.
	createSymlink := func(old, new string) { require.NoError(t, os.Symlink(old, new)) }

	// The resulting structure (full names were truncated for readability):
	//
	//                    tmpDir                                 symlinks
	//                  /    |    \                             /        \
	//                 /     |     \                           /          \
	//          more-yaml  s.yaml  empty-dir       more-yaml-symlink   slo-symlink.yml
	//         /    |    \
	//        /     |     \
	//       /      |      \
	//  saa.yaml  pad.yml  even-more-definitions
	//                    /     |       |   	\
	//                   /      |       | 		 \
	//               run.sh  a.yaml   k8s.yaml   p.json
	//
	// tmpDir:
	createFile("slo.yaml")
	createDir("empty-dir")
	// tmpDir/more-yaml:
	createDir("more-yaml")
	createFile("more-yaml/projects_and_direct.yml")
	createFile("more-yaml/service_and_agent.yaml")
	// tmpDir/more-yaml/even-more-definitions:
	createDir("more-yaml/even-more-definitions")
	createFile("more-yaml/even-more-definitions/annotations.yaml")
	createFile("more-yaml/even-more-definitions/k8s.yaml")
	createFile("more-yaml/even-more-definitions/run.sh")
	createFile("more-yaml/even-more-definitions/project.json")
	// symlinks:
	createSymlink(tmpDir("more-yaml"), symlinkDir("more-yaml-symlink"))
	createSymlink(tmpDir("slo.yaml"), symlinkDir("slo-symlink.yml"))

	// Prepare expected files located in tmpDir.
	allNobl9TmpFiles := []expectedMeta{
		{Name: "slo", ManifestSrc: tmpDir("slo.yaml")},
		{Name: "service_and_agent", ManifestSrc: tmpDir("more-yaml/service_and_agent.yaml")},
		{Name: "projects_and_direct", ManifestSrc: tmpDir("more-yaml/projects_and_direct.yml")},
		{Name: "annotations", ManifestSrc: tmpDir("more-yaml/even-more-definitions/annotations.yaml")},
		{Name: "project", ManifestSrc: tmpDir("more-yaml/even-more-definitions/project.json")},
	}
	// Prepare expected files located in pkg/definitions/test_data.
	allNobl9RelFiles := []expectedMeta{
		{Name: "slo", ManifestSrc: workingDir("test_data/inputs/slo.yaml")},
		{Name: "service_and_agent", ManifestSrc: workingDir("test_data/inputs/service_and_agent.yaml")},
		{Name: "projects_and_direct", ManifestSrc: workingDir("test_data/inputs/projects_and_direct.yml")},
		{Name: "annotations", ManifestSrc: workingDir("test_data/inputs/annotations.yaml")},
		{Name: "project", ManifestSrc: workingDir("test_data/inputs/project.json")},
	}

	const organization = "my-org"
	for name, test := range map[string]struct {
		Sources  []RawSource
		Expected []expectedMeta
	}{
		"read single file by name": {
			Sources:  []RawSource{tmpDir("slo.yaml")},
			Expected: []expectedMeta{{Name: "slo", ManifestSrc: tmpDir("slo.yaml")}},
		},
		"multiple single file sources by name": {
			Sources: []RawSource{tmpDir("slo.yaml"), tmpDir("more-yaml/service_and_agent.yaml")},
			Expected: []expectedMeta{
				{Name: "slo", ManifestSrc: tmpDir("slo.yaml")},
				{Name: "service_and_agent", ManifestSrc: tmpDir("more-yaml/service_and_agent.yaml")},
			},
		},
		"read immediate directory files with a dot": {
			Sources:  []RawSource{tmpDir(".")},
			Expected: []expectedMeta{{Name: "slo", ManifestSrc: tmpDir("slo.yaml")}},
		},
		"read immediate directory files with a wildcard": {
			Sources:  []RawSource{tmpDir("*")},
			Expected: []expectedMeta{{Name: "slo", ManifestSrc: tmpDir("slo.yaml")}},
		},
		"read all the files starting with 'slo'": {
			Sources:  []RawSource{tmpDir("**/slo*")},
			Expected: []expectedMeta{{Name: "slo", ManifestSrc: tmpDir("slo.yaml")}},
		},
		"read directory files with a glob pattern": {
			Sources:  []RawSource{tmpDir("*/*.yml")},
			Expected: []expectedMeta{{Name: "projects_and_direct", ManifestSrc: tmpDir("more-yaml/projects_and_direct.yml")}},
		},
		"read test_data directory files with a relative path": {
			Sources:  []RawSource{"test_data/inputs"},
			Expected: allNobl9RelFiles,
		},
		"read a single directory by name": {
			Sources: []RawSource{tmpDir("more-yaml/even-more-definitions")},
			Expected: []expectedMeta{
				{Name: "annotations", ManifestSrc: tmpDir("more-yaml/even-more-definitions/annotations.yaml")},
				{Name: "project", ManifestSrc: tmpDir("more-yaml/even-more-definitions/project.json")},
			},
		},
		"recurse the whole FS tree with a wildcard": {
			Sources:  []RawSource{tmpDir("**")},
			Expected: allNobl9TmpFiles,
		},
		"recurse the whole relative FS tree with a wildcard": {
			Sources:  []RawSource{workingDir("test_data/inputs/**")},
			Expected: allNobl9RelFiles,
		},
		"double wildcard inside the pattern": {
			Sources: []RawSource{tmpDir("**/even-more-definitions/*")},
			Expected: []expectedMeta{
				{Name: "annotations", ManifestSrc: tmpDir("more-yaml/even-more-definitions/annotations.yaml")},
				{Name: "project", ManifestSrc: tmpDir("more-yaml/even-more-definitions/project.json")},
			},
		},
		"duplicated sources with the same content are allowed": {
			Sources:  []RawSource{tmpDir("slo.yaml"), tmpDir(".")},
			Expected: []expectedMeta{{Name: "slo", ManifestSrc: tmpDir("slo.yaml")}},
		},
		"read a symlink to file": {
			Sources:  []RawSource{symlinkDir("slo-symlink.yml")},
			Expected: []expectedMeta{{Name: "slo", ManifestSrc: symlinkDir("slo-symlink.yml")}},
		},
		"read a symlink to directory with a wildcard": {
			Sources: []RawSource{symlinkDir("more-yaml-symlink/*")},
			Expected: []expectedMeta{
				{Name: "service_and_agent", ManifestSrc: symlinkDir("more-yaml-symlink/service_and_agent.yaml")},
				{Name: "projects_and_direct", ManifestSrc: symlinkDir("more-yaml-symlink/projects_and_direct.yml")},
			},
		},
		"read all directory symlinks through double wildcard": {
			Sources: []RawSource{symlinkDir("**")},
			Expected: []expectedMeta{
				{Name: "slo", ManifestSrc: symlinkDir("slo-symlink.yml")},
				{Name: "service_and_agent", ManifestSrc: symlinkDir("more-yaml-symlink/service_and_agent.yaml")},
				{Name: "projects_and_direct", ManifestSrc: symlinkDir("more-yaml-symlink/projects_and_direct.yml")},
				{Name: "annotations", ManifestSrc: symlinkDir("more-yaml-symlink/even-more-definitions/annotations.yaml")},
				{Name: "project", ManifestSrc: symlinkDir("more-yaml-symlink/even-more-definitions/project.json")},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			definitions, err := Read(ctx, MetadataAnnotations{Organization: organization}, test.Sources...)
			require.NoError(t, err)

			definitionsMatchExpected(t, definitions, test.Expected...)
		})
	}

	for name, test := range map[string]struct {
		Sources  []RawSource
		Expected error
	}{
		"missing file pattern for wildcard directory": {
			Sources:  []RawSource{tmpDir("**/even-more-definitions")},
			Expected: ErrNoFilesMatchingPattern,
		},
		"no files found under selected directory": {
			Sources:  []RawSource{tmpDir("empty-dir")},
			Expected: ErrNoFilesInPath,
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err = Read(ctx, MetadataAnnotations{Organization: organization}, test.Sources...)
			require.Error(t, err)
			assert.ErrorIs(t, err, test.Expected)
		})
	}
}

type expectedMeta struct {
	Name         string
	Organization string
	ManifestSrc  string
}

func definitionsMatchExpected(t *testing.T, definitions []sdk.AnyJSONObj, meta ...expectedMeta) {
	t.Helper()
	expected := make([]sdk.AnyJSONObj, 0, len(definitions))
	for _, m := range meta {
		if len(m.Organization) == 0 {
			m.Organization = "my-org"
		}
		buf := bytes.NewBuffer([]byte{})
		err := templates.ExecuteTemplate(buf, m.Name+".tpl.json", m)
		require.NoError(t, err)
		var decoded interface{}
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		switch v := decoded.(type) {
		case []interface{}:
			for _, i := range v {
				expected = append(expected, i.(map[string]interface{}))
			}
		case map[string]interface{}:
			expected = append(expected, v)
		}
	}
	require.Equal(t, len(expected), len(definitions))

	assert.ElementsMatch(t, expected, definitions)
}

// readTestFile attempts to read the designated file from test_data folder.
func readTestFile(t *testing.T, filename string) *bytes.Buffer {
	t.Helper()
	data, err := testData.ReadFile(filepath.Join("test_data", "inputs", filename))
	require.NoError(t, err)
	return bytes.NewBuffer(data)
}
