package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/nobl9/nobl9-go/internal/pathutils"
	"github.com/nobl9/nobl9-go/manifest"
	v1alphaExamples2 "github.com/nobl9/nobl9-go/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/sdk"
)

type examplesGeneratorConfig struct {
	Examples []v1alphaExamples2.Example
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
			Examples: v1alphaExamples2.Labels(),
			Path:     filepath.Join(basePath, "labels_examples.yaml"),
		},
		{
			Examples: v1alphaExamples2.MetadataAnnotations(),
			Path:     filepath.Join(basePath, "metadata_annotations_examples.yaml"),
		},
	}
	// Standard examples.
	allExamples := [][]v1alphaExamples2.Example{
		v1alphaExamples2.Project(),
		v1alphaExamples2.Service(),
		v1alphaExamples2.AlertMethod(),
		v1alphaExamples2.SLO(),
		v1alphaExamples2.Agent(),
		v1alphaExamples2.Direct(),
		v1alphaExamples2.AlertPolicy(),
		v1alphaExamples2.AlertSilence(),
		v1alphaExamples2.Annotation(),
		v1alphaExamples2.BudgetAdjustment(),
		v1alphaExamples2.DataExport(),
		v1alphaExamples2.RoleBinding(),
		v1alphaExamples2.Report(),
	}
	for _, examples := range allExamples {
		object := examples[0].GetObject().(manifest.Object)
		basePath := filepath.Join(
			manifestPath,
			object.GetVersion().VersionString(),
			object.GetKind().ToLower(),
		)
		grouped := groupBy(examples, func(e v1alphaExamples2.Example) string { return e.GetVariant() })
		for variant, examples := range grouped {
			var path string
			if len(grouped) == 1 {
				// If we don't have any variants, we can write all examples into examples.yaml file.
				path = filepath.Join(basePath, "examples.yaml")
			} else {
				path = filepath.Join(basePath, "examples", strings.ReplaceAll(strings.ToLower(variant), " ", "-")+".yaml")
			}
			config := examplesGeneratorConfig{
				Examples: examples,
				Path:     path,
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
