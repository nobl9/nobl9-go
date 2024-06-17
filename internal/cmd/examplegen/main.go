package examplegen

import (
	"path/filepath"

	"github.com/goccy/go-yaml"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/manifest"
)

type examplesGeneratorFunc func() any

type examplesGeneratorConfig struct {
	Generate examplesGeneratorFunc
	Path     string
}

const manifestPath = "manifest"

func main() {
	configs := getV1alphaExamplesRegistry()
	for _, config := range configs {
		v := config.Generate()
		if object, ok := v.(manifest.Object); ok && config.Path != "" {
			config.Path = filepath.Join(
				manifestPath,
				object.GetVersion().VersionString(),
				object.GetKind().ToLower(),
				"examples.yaml",
			)
		}

		enc := yaml.NewEncoder()
	}
}

func getV1alphaExamplesRegistry() []examplesGeneratorConfig {
	path := filepath.Join(manifestPath, "v1alpha")
	return []examplesGeneratorConfig{
		{Generate: generify(v1alphaExamples.Project)},
		{Generate: generify(v1alphaExamples.Service)},
		{
			Generate: generify(v1alphaExamples.Labels),
			Path:     filepath.Join(path, "labels_examples.yaml"),
		},
		{
			Generate: generify(v1alphaExamples.MetadataAnnotations),
			Path:     filepath.Join(path, "metadata_annotations_examples.yaml"),
		},
	}
}

func generify[T any](generator func() T) examplesGeneratorFunc {
	return func() any { return generator() }
}
