package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	v1alphaExamples "github.com/nobl9/nobl9-go/internal/manifest/v1alpha/examples"
	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/manifest/v1alpha"
	v1alphaSLO "github.com/nobl9/nobl9-go/manifest/v1alpha/slo"
	"github.com/nobl9/nobl9-go/sdk"
)

type generatedSnippet struct {
	Name    string `json:"name"`
	Snippet string `json:"snippet"`
}

func main() {
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
		v1alphaExamples.Report(),
	}

	generatedList := make([]generatedSnippet, 0, len(allExamples))
	for _, examples := range allExamples {
		for _, example := range examples {
			object, ok := example.GetObject().(manifest.Object)
			if !ok {
				continue
			}
			// For SLOs we only want to get examples for certain configurations.
			if _, isSLO := object.(v1alphaSLO.SLO); isSLO && !strings.HasSuffix(
				example.GetSubVariant(),
				"good over total SLO using Occurrences budgeting method and Rolling time window",
			) {
				continue
			}
			genericObject, err := formatObject(object)
			if err != nil {
				panic(err)
			}
			var buf bytes.Buffer
			if err = sdk.EncodeObject(genericObject, &buf, manifest.ObjectFormatYAML); err != nil {
				panic(err)
			}
			var name string
			if example.GetVariant() != "" {
				name = fmt.Sprintf("%s %s", example.GetVariant(), object.GetKind())
			} else if example.GetSubVariant() != "" && !slices.Contains([]manifest.Kind{manifest.KindSLO, manifest.KindAgent, manifest.KindDirect}, object.GetKind()) {
				name = fmt.Sprintf("%s %s", example.GetSubVariant(), object.GetKind())
			} else {
				name = object.GetKind().String()
			}
			generatedList = append(generatedList, generatedSnippet{
				Name:    strings.ReplaceAll(strings.ToLower(name), "-", ""),
				Snippet: buf.String(),
			})
		}
	}
	unique := make(map[string]bool)
	uniqueList := make([]generatedSnippet, 0, len(generatedList))
	for _, snippet := range generatedList {
		if unique[snippet.Name] {
			fmt.Fprintf(os.Stderr, "Duplicate snippet name: %s (dropping)\n", snippet.Name)
			continue
		}
		unique[snippet.Name] = true
		uniqueList = append(uniqueList, snippet)
	}
	data, err := json.Marshal(uniqueList)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile("snippets.json", data, 0o600); err != nil {
		panic(err.Error())
	}
}

func formatObject(object manifest.Object) (manifest.Object, error) {
	data, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}
	var generic v1alpha.GenericObject
	if err = json.Unmarshal(data, &generic); err != nil {
		return nil, err
	}

	placeholderIndex := 1
	getPlaceholder := func() string {
		return fmt.Sprintf("$%d", placeholderIndex)
	}

	metadata := generic["metadata"].(map[string]any)
	metadata["name"] = getPlaceholder()
	if metadata["displayName"] != nil {
		metadata["displayName"] = getPlaceholder()
	}
	placeholderIndex++
	if metadata["project"] != nil {
		metadata["project"] = getPlaceholder()
		placeholderIndex++
	}
	// Simplify the snippets.
	delete(metadata, "labels")
	delete(metadata, "annotations")
	generic["metadata"] = metadata

	spec := generic["spec"].(map[string]any)
	if spec["description"] != nil {
		spec["description"] = getPlaceholder()
		placeholderIndex++
	}
	return generic, nil
}
