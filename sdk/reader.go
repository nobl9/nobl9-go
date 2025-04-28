package sdk

import (
	"cmp"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
)

const APIVersionRegex = `"?apiVersion"?\s*:\s*"?n9`

// RawObjectSource may be interpreted as:
//   - file path as [ObjectSourceTypeFile] or [ObjectSourceTypeDirectory]
//   - glob pattern as [ObjectSourceTypeGlobPattern]
//   - URL as [ObjectSourceTypeURL]
//   - input provided via [io.Reader], like [os.Stdin] as [ObjectSourceTypeReader]
type RawObjectSource = string

type (
	// RawDefinition stores both the resolved source and raw resource definition.
	RawDefinition struct {
		// SourceType is the original [ObjectSource.Type].
		SourceType ObjectSourceType
		// ResolvedSource is a single definition's source descriptor.
		// For instance, if the definition was read by applying glob pattern to [ResolveObjectSources],
		// the [ResolvedSource] will be a specific file path.
		ResolvedSource string
		// Definition is the raw bytes content of the resource source.
		Definition []byte
	}
)

// ReadObjects resolves the [RawObjectSource] it receives
// and reads all [manifest.Object] from the resolved [ObjectSource].
//
// Refer to [ReadObjectsFromSources] for more details on the objects' reading logic.
func ReadObjects(ctx context.Context, rawSources ...RawObjectSource) ([]manifest.Object, error) {
	sources, err := ResolveObjectSources(rawSources...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve all raw sources")
	}
	return ReadObjectsFromSources(ctx, sources...)
}

const unknownSource = "-"

// ReadObjectsFromSources reads from the provided [ObjectSource] based on the [ObjectSourceType].
// It calculates a sum for each definition read from [ObjectSource] and won't create duplicates.
// This allows the user to combine [ObjectSource] with potentially overlapping paths.
// If the same exact definition is identified with multiple sources,
// it will pick the first [ObjectSource] path it encounters.
//
// If the [ObjectSource] is of type [ObjectSourceTypeGlobPattern] or [ObjectSourceTypeDirectory]
// and a file does not contain the required [APIVersionRegex], it is skipped.
// However, in case of [ObjectSourceTypeFile], it will throw [ErrInvalidFile] error.
//
// Each [ObjectSourceType] is handled according to the following logic:
//
//  1. [ObjectSourceTypeFile] and [ObjectSourceTypeDirectory]
//     The path can point to a single file or a directory.
//     If it's a directory, all files with the supported extension will be read.
//
//  2. [ObjectSourceTypeGlobPattern]
//     All files matching the pattern will be read.
//     On top of what is supported by [filepath.Match],
//     the pattern may contain double star '**' wildcard placeholders.
//     The double start will be interpreted as a recursive directory search.
//
//  3. [ObjectSourceTypeURL]
//     The URL to fetch object definitions from.
//     The endpoint at the provided URL should handle GET request by responding
//     with status code 200 and JSON or YAML encoded representation of [manifest.Object].
//
//     Note: This URL is not designed to fetch [manifest.Object] from the Nobl9 API.
//     It can be used, for instance, to fetch the objects from the users internal repository.
//     In order to read [manifest.Object] from the Nobl9 API, use [Client.Objects] API.
//
//  4. [ObjectSourceTypeReader]
//     The [ObjectSource.Reader] is read directly and [ObjectSource.Paths] is ignored.
func ReadObjectsFromSources(ctx context.Context, sources ...*ObjectSource) ([]manifest.Object, error) {
	definitions, err := ReadRawDefinitionsFromSources(ctx, sources...)
	if err != nil {
		return nil, err
	}
	definitions, err = filterRawDefinitions(definitions)
	if err != nil {
		return nil, err
	}
	return processRawDefinitions(definitions)
}

// ReadRawDefinitionsFromSources is a low level function which allows reading
// resource definitions from a list of resolved [ObjectSource].
// For more details refer to [ReadObjectsFromSources] docs.
func ReadRawDefinitionsFromSources(ctx context.Context, sources ...*ObjectSource) ([]*RawDefinition, error) {
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Raw > sources[j].Raw
	})
	uniqueDefinitions := make(map[string]*RawDefinition, len(sources))
	var (
		err error
		def []byte
	)
	for _, src := range sources {
		if src.Type == ObjectSourceTypeReader {
			switch len(src.Paths) {
			case 0:
				src.Paths = []string{unknownSource}
			case 1:
				break
			default:
				return nil, ErrSourceTypeReaderPath
			}
		}
		for _, path := range src.Paths {
			switch src.Type {
			case ObjectSourceTypeReader:
				def, err = readFromReader(src.Reader)
			case ObjectSourceTypeURL:
				def, err = readFromURL(ctx, path)
			case ObjectSourceTypeDirectory, ObjectSourceTypeGlobPattern, ObjectSourceTypeFile:
				def, err = readFromFile(path)
			default:
				err = ErrInvalidSourceType
			}
			if err != nil {
				return nil, errors.Wrapf(err, "failed to read resource definitions from '%s'", src)
			}
			hash := getRawDefinitionHash(def)
			if _, srcExists := uniqueDefinitions[hash]; !srcExists {
				uniqueDefinitions[hash] = &RawDefinition{
					SourceType:     src.Type,
					ResolvedSource: path,
					Definition:     def,
				}
			}
		}
	}
	definitions := slices.SortedFunc(
		maps.Values(uniqueDefinitions),
		func(a, b *RawDefinition) int {
			if cmpSource := cmp.Compare(a.SourceType, b.SourceType); cmpSource != 0 {
				return cmpSource
			}
			return cmp.Compare(a.ResolvedSource, b.ResolvedSource)
		},
	)
	return definitions, nil
}

var (
	ErrIoReaderIsNil          = errors.New("io.Reader must no be nil")
	ErrNoFilesMatchingPattern = errors.Errorf(
		"no Nobl9 resource definition files matched the provided path pattern, %s", matchingRulesDisclaimer)
	ErrNoFilesInPath = errors.Errorf("no Nobl9 resource definition files were found under selected path, %s",
		matchingRulesDisclaimer)
	ErrInvalidFile = errors.Errorf("valid Nobl9 resource definition must match against the following regex: '%s'",
		APIVersionRegex)
	ErrInvalidSourceType    = errors.New("invalid ObjectSourceType provided")
	ErrSourceTypeReaderPath = errors.New(
		"ObjectSourceTypeReader ObjectSource may define at most a single ObjectSource.Path")

	matchingRulesDisclaimer = fmt.Sprintf(
		"valid resource definition file must have one of the extensions: [%s]",
		strings.Join(supportedFileExtensions, ","))
)

func getRawDefinitionHash(def []byte) string {
	sum := sha256.Sum256(def)
	return string(sum[:])
}

func readFromReader(in io.Reader) ([]byte, error) {
	if in == nil {
		return nil, ErrIoReaderIsNil
	}
	return io.ReadAll(in)
}

func filterRawDefinitions(definitions []*RawDefinition) ([]*RawDefinition, error) {
	filtered := make([]*RawDefinition, 0, len(definitions))
	for _, def := range definitions {
		// nolint: exhaustive
		switch def.SourceType {
		case ObjectSourceTypeFile:
			if !apiVersionRegex.Match(def.Definition) {
				return nil, ErrInvalidFile
			}
		case ObjectSourceTypeDirectory, ObjectSourceTypeGlobPattern:
			if !apiVersionRegex.Match(def.Definition) {
				continue
			}
		}
		filtered = append(filtered, def)
	}
	return filtered, nil
}

// TODO: in the future if we'd run sloctl daemon or web server, this should become a pool instead.
// HTTP clients should be reused whenever possible as they cache TCP connections, they are also
// concurrently safe by design.
// The factory is defined in a package variable to allow testing of HTTPS requests with httptest package.
var httpClientFactory = func(url string) *http.Client {
	return newRetryableHTTPClient(10*time.Second, nil)
}

func readFromURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating new GET %s request", url)
	}
	resp, err := httpClientFactory(url).Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error receiving GET %s response", url)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("GET %s response: %d %s", url, resp.StatusCode, string(data))
	}
	return io.ReadAll(resp.Body)
}

func readFromFile(fp string) ([]byte, error) {
	// #nosec G304
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s file", fp)
	}
	return data, nil
}
