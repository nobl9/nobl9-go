package definitions

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/nobl9/nobl9-go/manifest"
	"github.com/nobl9/nobl9-go/sdk/retryhttp"
)

type (
	// RawSource may be interpreted as (with interpretation):
	// - file path  (SourceTypeFile or SourceTypeDirectory)
	// - glob pattern (SourceTypeGlobPattern)
	// - URL (SourceTypeURL)
	// - input provided via io.Reader, like os.Stdin (SourceTypeReader)
	RawSource = string

	// rawDefinition stores both the resolved source and raw resource definition.
	rawDefinition struct {
		// ResolvedSource
		ResolvedSource string
		Definition     []byte
	}
	// rawDefinitions simulates a set, map of unique resource definitions.
	// Uniqueness is calculated on all bytes via SHA256 sum.
	rawDefinitions = map[ /* raw definition hash */ string]rawDefinition
)

// Read resolves the RawSource(s) it receives and calls ReadSources on the resolved Source(s).
func Read(ctx context.Context, rawSources ...RawSource) ([]manifest.Object, error) {
	sources, err := ResolveSources(rawSources...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve all raw sources")
	}
	return ReadSources(ctx, sources...)
}

const unknownSource = "-"

// ReadSources reads from the provided Source(s) based on the SourceType.
// For SourceTypeReader it will read directly from Source.Reader,
// otherwise it reads from all the Source.Paths. It calculates a sum for
// each definition read from Source and won't create duplicates. This
// allows the user to combine Source(s) with possibly overlapping paths.
// If the same exact definition is identified with multiple sources, it
// will choose the first Source path it encounters. If the Source is of
// type SourceTypeGlobPattern or SourceTypeDirectory and a file does not
// contain the required APIVersionRegex, it is skipped. However in case
// of SourceTypeFile, it will thrown ErrInvalidFile error.
func ReadSources(ctx context.Context, sources ...*Source) ([]manifest.Object, error) {
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Raw > sources[j].Raw
	})
	definitions := make(rawDefinitions, len(sources))
	var (
		err error
		def []byte
	)
	for _, src := range sources {
		if src.Type == SourceTypeReader {
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
			case SourceTypeReader:
				def, err = readFromReader(src.Reader)
			case SourceTypeURL:
				def, err = readFromURL(ctx, path)
			case SourceTypeDirectory, SourceTypeGlobPattern:
				def, err = readFromFile(path)
				// We only want to fail on the regex check when a single file is supplied.
				if errors.Is(err, ErrInvalidFile) {
					continue
				}
			case SourceTypeFile:
				def, err = readFromFile(path)
			default:
				err = ErrInvalidSourceType
			}
			if err != nil {
				return nil, errors.Wrapf(err, "failed to read resource definitions from '%s'", src)
			}
			appendUniqueDefinition(definitions, path, def)
		}
	}
	return processRawDefinitions(definitions)
}

var (
	ErrIoReaderIsNil          = errors.New("io.Reader must no be nil")
	ErrNoFilesMatchingPattern = errors.Errorf(
		"no Nobl9 resource definition files matched the provided path pattern, %s", matchingRulesDisclaimer)
	ErrNoFilesInPath = errors.Errorf("no Nobl9 resource definition files were found under selected path, %s",
		matchingRulesDisclaimer)
	ErrInvalidFile = errors.Errorf("valid Nobl9 resource definition must match against the following regex: '%s'",
		APIVersionRegex)
	ErrInvalidSourceType    = errors.New("invalid SourceType provided")
	ErrSourceTypeReaderPath = errors.New("SourceTypeReader Source may define at most a single Source.Path")

	matchingRulesDisclaimer = fmt.Sprintf(
		"valid resource definition file must have one of the extensions: [%s]",
		strings.Join(supportedFileExtensions, ","))
)

func appendUniqueDefinition(defs rawDefinitions, src string, def []byte) {
	sum := sha256.Sum256(def)
	hash := string(sum[:])
	if _, srcExists := defs[hash]; srcExists {
		return
	}
	defs[hash] = rawDefinition{ResolvedSource: src, Definition: def}
}

func readFromReader(in io.Reader) ([]byte, error) {
	if in == nil {
		return nil, ErrIoReaderIsNil
	}
	return io.ReadAll(in)
}

// TODO: in the future if we'd run sloctl daemon or web server, this should become a pool instead.
// HTTP clients should be reused whenever possible as they cache TCP connections, they are also
// concurrently safe by design.
// The factory is defined in a package variable to allow testing of HTTPS requests with httptest package.
var httpClientFactory = func(url string) *http.Client {
	return retryhttp.NewClient(10*time.Second, nil)
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

const APIVersionRegex = `"?apiVersion"?\s*:\s*"?n9`

var apiVersionRegex = regexp.MustCompile(APIVersionRegex)

// #nosec G304
func readFromFile(fp string) ([]byte, error) {
	data, err := os.ReadFile(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s file", fp)
	}
	// The exact version is not provided as it might change.
	// The n9 prefix however is not likely to ever change.
	// Since the version is always at the top of the document bytes.Contain
	// should quickly find the first match.
	if !apiVersionRegex.Match(data) {
		return nil, ErrInvalidFile
	}
	return data, nil
}
