package definitions

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ResolveSources calls ResolveSource on all supplied RawSource(s) and aggregates the resolved Source(s).
// It fails fast on the first encountered error.
func ResolveSources(rawSources ...RawSource) ([]*Source, error) {
	sources := make([]*Source, 0, len(rawSources))
	for _, raw := range rawSources {
		src, err := ResolveSource(raw)
		if err != nil {
			return nil, err
		}
		sources = append(sources, src)
	}
	return sources, nil
}

// ResolveSource attempts to resolve a single RawSource producing a Source instance read to be passed to ReadSources.
// It interprets the provided URI and associates it with a specific SourceType.
// If you wish to create a SourceTypeReader Source you should use a separate method: NewReaderSource.
func ResolveSource(rawSource RawSource) (src *Source, err error) {
	src = &Source{Raw: rawSource}
	switch {
	case hasURLSchema(rawSource):
		src.Type = SourceTypeURL
		src.Paths = []string{rawSource}
	case hasGlobMeta(rawSource):
		src.Type = SourceTypeGlobPattern
		src.Paths, err = resolveGlobPattern(rawSource)
	default:
		src.Type, src.Paths, err = resolveFSPath(rawSource)
	}
	return src, err
}

// NewReaderSource creates a special instance of Source with SourceTypeReader.
// ReadSources will process the Source by reading form the provided io.Reader.
func NewReaderSource(r io.Reader, source RawSource) *Source {
	return &Source{
		Type:   SourceTypeReader,
		Paths:  []string{source},
		Reader: r,
		Raw:    source,
	}
}

// Source represents a single resource definition source.
type Source struct {
	// Type defines how the Source should be read when passed to ReadSources.
	Type SourceType
	// Paths lists all resolved URIs the Source points at.
	Paths []string
	// Reader may be optionally provided with SourceTypeReader for ReadSources to read from the io.Reader.
	Reader io.Reader
	// Raw is the original, unresolved RawSource, an example might be a relative path
	// which was resolved to its absolute form.
	Raw RawSource
}

type SourceType int

const (
	SourceTypeFile SourceType = iota
	SourceTypeDirectory
	SourceTypeGlobPattern
	SourceTypeURL
	SourceTypeReader
)

func (s Source) String() string {
	return fmt.Sprintf("{SourceType: %s, Raw: %s}", s.Type, s.Raw)
}

var supportedFileExtensions = []string{".yaml", ".yml", ".json"}

// GetSupportedFileExtensions returns the file extensions which are used to filter out files to be processed.
func GetSupportedFileExtensions() []string {
	s := make([]string, len(supportedFileExtensions))
	copy(s, supportedFileExtensions)
	return s
}

// resolveFSPath we'll recognize if the provided path points to a single file
// or a directory. If it's a directory it will resolve paths of all its
// immediate children (1st level) and will not recurse it's structure.
func resolveFSPath(path string) (typ SourceType, paths []string, err error) {
	path, err = filepath.Abs(filepath.Clean(path))
	if err != nil {
		return -1, nil, err
	}
	stat, err := os.Stat(path)
	if err != nil {
		return -1, nil, err
	}
	switch {
	case stat.IsDir():
		de, err := os.ReadDir(path)
		if err != nil {
			return -1, nil, err
		}
		paths = make([]string, 0, len(de))
		for _, e := range de {
			fp := filepath.Join(path, e.Name())
			if e.IsDir() || !hasSupportedFileExtension(fp) {
				continue
			}
			paths = append(paths, fp)
		}
		if len(paths) == 0 {
			return -1, nil, ErrNoFilesInPath
		}
		typ = SourceTypeDirectory
	default:
		typ = SourceTypeFile
		paths = []string{path}
	}
	return typ, paths, nil
}

// resolveGlobPattern evaluates patterns defined in filepath.Match documentation.
// It supports double '**' wildcards directory expansion via a 3rd party library:
// https://github.com/bmatcuk/doublestar. The library is a pure (no dependencies)
// implementation of glob patterns with double wildcard support.
// Whenever a double wildcard is detected doublestar.Glob is used, otherwise we
// use filepath.Glob to minimize the impact of potential bugs in doublestar.Glob.
// The reasons for Go's lack of '**' support are outlined in this issue:
// https://github.com/golang/go/issues/11862.
func resolveGlobPattern(path string) (paths []string, err error) {
	path, err = filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	switch {
	case strings.Contains(path, "**"):
		// Call to filepath.ToSlash makes it platform independent since
		// doublestar.Glob is meant as a drop-in replacement for fs.Glob
		// which relies on '/' separator in the path.
		path = filepath.ToSlash(path)
		var base string
		base, path = doublestar.SplitPattern(path)
		paths, err = doublestar.Glob(
			os.DirFS(base),
			path,
			doublestar.WithFilesOnly(),
			doublestar.WithFailOnPatternNotExist())
		// Unlike filepath.Glob, doublestar.Glob operates on provided
		// filesystem, which is the base directory. We need to append it
		// afterwards to keep the absolute path.
		for i := range paths {
			paths[i] = filepath.Join(base, paths[i])
		}
	default:
		paths, err = filepath.Glob(path)
	}
	if err != nil {
		return nil, err
	}
	filteredPaths := make([]string, 0, len(paths))
	for i := range paths {
		if !hasSupportedFileExtension(paths[i]) {
			continue
		}
		// To keep it platform independent we need to make sure the RawSource
		// has the correct path separators. If we used doublestar.Glob we
		// replaced Windows path separator with '/', so we need to roll it back.
		filteredPaths = append(filteredPaths, filepath.FromSlash(paths[i]))
	}
	if len(filteredPaths) == 0 {
		return nil, ErrNoFilesMatchingPattern
	}
	return filteredPaths, nil
}

// hasSupportedFileExtension checks if we're dealing with YAML or JSON file comparing file extension suffix.
// It's faster to do the simple comparison strings.HasSuffix does then call filepath.Ext.
func hasSupportedFileExtension(fp string) bool {
	ext := filepath.Ext(fp)
	for i := range supportedFileExtensions {
		if ext == supportedFileExtensions[i] {
			return true
		}
	}
	return false
}

// hasURLSchema performs a trivial prefix check for the presence of either 'http://' or 'https://'.
// It does not verify if the path is a valid URL, it will also not allow otherwise valid URLs
// since a specific schema subset is required.
func hasURLSchema(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// hasGlobMeta reports whether path contains any of the magic characters recognized by filepath.Match.
// Copied from filepath/match.go.
func hasGlobMeta(path string) bool {
	magicChars := `*?[`
	if runtime.GOOS != "windows" {
		magicChars = `*?[\`
	}
	return strings.ContainsAny(path, magicChars)
}
