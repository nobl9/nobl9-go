package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/internal/pathutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk"
)

type examplesGeneratorConfig struct {
	Examples []v1alphaExamples.Example
	Path     string
	Comments yaml.CommentMap
}

const manifestPath = "manifest"

func main() {
	rootPath := pathutils.FindModuleRoot()
	configs := getV1alphaExamplesConfigs()
	for _, config := range configs {
		examples := make([]any, 0, len(config.Examples))
		for _, variant := range config.Examples {
			examples = append(examples, variant.GetObject())
		}
		config.Path = filepath.Join(rootPath, config.Path)
		var input any
		if len(examples) == 1 {
			input = examples[0]
		} else {
			input = examples
		}
		if err := writeExamples(input, config.Path, config.Comments); err != nil {
			panic(err.Error())
		}
	}
}

func getV1alphaExamplesConfigs() []examplesGeneratorConfig {
	basePath := filepath.Join(manifestPath, "v1alpha")
	// Non-standard examples.
	configs := []examplesGeneratorConfig{
		{
			Examples: v1alphaExamples.Labels(),
			Path:     filepath.Join(basePath, "labels_examples.yaml"),
		},
		{
			Examples: v1alphaExamples.MetadataAnnotations(),
			Path:     filepath.Join(basePath, "metadata_annotations_examples.yaml"),
		},
	}
	// Standard examples.
	allExamples := [][]v1alphaExamples.Example{
		v1alphaExamples.Project(),
		v1alphaExamples.Service(),
		v1alphaExamples.AlertMethod(),
		v1alphaExamples.SLO(),
		v1alphaExamples.Agent(),
		v1alphaExamples.Direct(),
		v1alphaExamples.AlertPolicy(),
		v1alphaExamples.AlertSilence(),
		v1alphaExamples.Annotation(),
		v1alphaExamples.BudgetAdjustment(),
		v1alphaExamples.DataExport(),
		v1alphaExamples.RoleBinding(),
	}
	for _, examples := range allExamples {
		object := examples[0].GetObject().(manifest.Object)
		basePath := filepath.Join(
			manifestPath,
			object.GetVersion().VersionString(),
			object.GetKind().ToLower(),
		)
		grouped := groupBy(examples, func(e v1alphaExamples.Example) string { return e.GetVariant() })
		// If we don't have any variants, we can write all examples into examples.yaml file.
		if len(grouped) == 1 {
			configs = append(configs, examplesGeneratorConfig{
				Examples: examples,
				Path:     filepath.Join(basePath, "examples.yaml"),
			})
			continue
		}
		for variant, examples := range grouped {
			config := examplesGeneratorConfig{
				Examples: examples,
				Path:     filepath.Join(basePath, "examples", strings.ReplaceAll(strings.ToLower(variant), " ", "-")+".yaml"),
				Comments: make(yaml.CommentMap),
			}
			if len(examples) == 1 {
				configs = append(configs, config)
				continue
			}
			if examples[0].GetSubVariant() != "" {
				sort.Slice(examples, func(i, j int) bool {
					return examples[i].GetSubVariant() < examples[j].GetSubVariant()
				})
			}
			for i, example := range examples {
				comments := example.GetYAMLComments()
				if len(comments) == 0 {
					continue
				}
				for i := range comments {
					comments[i] = " " + comments[i]
				}
				config.Comments[fmt.Sprintf("$[%d]", i)] = []*yaml.Comment{yaml.HeadComment(comments...)}
			}
			configs = append(configs, config)
		}
	}
	return configs
}

func writeExamples(v any, path string, comments yaml.CommentMap) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	// #nosec G304
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if object, ok := v.(manifest.Object); ok {
		return sdk.EncodeObject(object, file, manifest.ObjectFormatYAML)
	}
	opts := []yaml.EncodeOption{
		yaml.Indent(2),
		yaml.UseLiteralStyleIfMultiline(true),
	}
	if len(comments) > 0 {
		opts = append(opts, yaml.WithComment(comments))
	}
	enc := yaml.NewEncoder(file, opts...)
	return enc.Encode(v)
}

func groupBy[K comparable, V any](s []V, key func(V) K) map[K][]V {
	m := make(map[K][]V)
	for _, v := range s {
		k := key(v)
		m[k] = append(m[k], v)
	}
	return m
}
