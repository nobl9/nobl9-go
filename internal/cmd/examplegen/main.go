package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/internal/pathutils"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
)

type examplesGeneratorFunc func() any

type examplesGeneratorConfig struct {
	Generate examplesGeneratorFunc
	Path     string
	Comments yaml.CommentMap
}

const manifestPath = "manifest"

func main() {
	rootPath := pathutils.FindModuleRoot()
	configs := getV1alphaExamplesConfigs()
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
		if err := writeExamples(v, config.Path, config.Comments); err != nil {
			errFatal(err.Error())
		}
	}
}

func writeExamples(v any, path string, comments yaml.CommentMap) error {
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

func getV1alphaExamplesConfigs() []examplesGeneratorConfig {
	path := filepath.Join(manifestPath, "v1alpha")
	config := []examplesGeneratorConfig{
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
	config = append(config, getV1alphaSLOExamplesConfigs(path)...)
	return config
}

func getV1alphaSLOExamplesConfigs(path string) []examplesGeneratorConfig {
	variantsPerDataSource := make(
		map[v1alpha.DataSourceType][]v1alphaExamples.SLOVariant,
		len(v1alpha.DataSourceTypeValues()),
	)
	for _, variant := range v1alphaExamples.SLO() {
		variantsPerDataSource[variant.DataSourceType] = append(
			variantsPerDataSource[variant.DataSourceType],
			variant,
		)
	}
	config := make([]examplesGeneratorConfig, 0, len(variantsPerDataSource))
	for dataSourceType, variants := range variantsPerDataSource {
		comments := make(yaml.CommentMap, len(variants))
		for i, variant := range variants {
			texts := []string{
				fmt.Sprintf(" Metric type: %s", variant.MetricVariant),
				fmt.Sprintf(" Budgeting method: %s", variant.BudgetingMethod),
				fmt.Sprintf(" Time window type: %s", variant.TimeWindowType),
			}
			if variant.MetricSubVariant != "" {
				texts = slices.Insert(texts, 1, fmt.Sprintf(" Metric variant: %s", variant.MetricSubVariant))
			}
			comments[fmt.Sprintf("$[%d]", i)] = []*yaml.Comment{yaml.HeadComment(texts...)}
		}
		config = append(config, examplesGeneratorConfig{
			Generate: func() any {
				return mapSlice(variants, func(v v1alphaExamples.SLOVariant) v1alphaSLO.SLO { return v.SLO })
			},
			Path: filepath.Join(
				path,
				"slo",
				"examples",
				fmt.Sprintf("%s.yaml", strings.ToLower(dataSourceType.String())),
			),
			Comments: comments,
		})
	}
	return config
}

func generify[T any](generator func() T) examplesGeneratorFunc {
	return func() any { return generator() }
}

func errFatal(f string) {
	fmt.Fprintln(os.Stderr, f)
	os.Exit(1)
}

func mapSlice[T, N any](s []T, m func(T) N) []N {
	r := make([]N, len(s))
	for i, v := range s {
		r[i] = m(v)
	}
	return r
}
