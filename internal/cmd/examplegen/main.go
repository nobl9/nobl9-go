package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/internal/pathutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
)

type examplesGeneratorFunc func() any

type examplesGeneratorConfig struct {
	Generate examplesGeneratorFunc
	Path     string
}

const manifestPath = "manifest"

func main() {
	rootPath := pathutils.FindModuleRoot()
	configs := getV1alphaExamplesRegistry()
	for _, config := range configs {
		v := config.Generate()
		if object, ok := v.(manifest.Object); ok && config.Path == "" {
			config.Path = filepath.Join(
				manifestPath,
				object.GetVersion().VersionString(),
				object.GetKind().ToLower(),
				"examples.yaml",
			)
		}
		config.Path = filepath.Join(rootPath, config.Path)
		if err := writeExamples(v, config.Path); err != nil {
			errFatal(err.Error())
		}
	}
}

func writeExamples(v any, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if object, ok := v.(manifest.Object); ok {
		return sdk.PrintObject(object, file, manifest.ObjectFormatYAML)
	}
	enc := yaml.NewEncoder(file, yaml.Indent(2))
	return enc.Encode(v)
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

func errFatal(f string) {
	fmt.Fprintln(os.Stderr, f)
	os.Exit(1)
}
