package sdk

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ResolveObjectSources calls [ResolveObjectSource] on all supplied [RawObjectSource]
// and aggregates the resolved [ObjectSource].
// It fails fast on the first encountered error.
func ResolveObjectSources(rawSources ...RawObjectSource) ([]*ObjectSource, error) {
	sources := make([]*ObjectSource, 0, len(rawSources))
	for _, raw := range rawSources {
		src, err := ResolveObjectSource(raw)
		if err != nil {
			return nil, err
		}
		sources = append(sources, src)
	}
	return sources, nil
}

// ResolveObjectSource attempts to resolve a single [RawObjectSource] producing an [ObjectSource]
// instance read to be passed to [ReadObjectsFromSources].
// It interprets the provided URI and associates it with a specific [ObjectSourceType].
// If you wish to create an [ObjectSource] of type [ObjectSourceTypeReader]
// you should use a separate method: [NewObjectSourceReader].
func ResolveObjectSource(rawSource RawObjectSource) (src *ObjectSource, err error) {
	src = &ObjectSource{Raw: rawSource}
	switch {
	case hasURLSchema(rawSource):
		src.Type = ObjectSourceTypeURL
		src.Paths = []string{rawSource}
	case hasGlobMeta(rawSource):
		src.Type = ObjectSourceTypeGlobPattern
		src.Paths, err = resolveGlobPattern(rawSource)
	default:
		src.Type, src.Paths, err = resolveFSPath(rawSource)
	}
	return src, err
}

// NewObjectSourceReader creates a special instance of [ObjectSource] with [ObjectSourceTypeReader].
// [ReadObjectsFromSources] will process the ObjectSource by reading form the provided io.Reader.
func NewObjectSourceReader(r io.Reader, source RawObjectSource) *ObjectSource {
	return &ObjectSource{
		Type:   ObjectSourceTypeReader,
		Paths:  []string{source},
		Reader: r,
		Raw:    source,
	}
}

// ObjectSource represents a single resource definition source.
type ObjectSource struct {
	// Type defines how the [ObjectSource] should be read when passed to [ReadObjectsFromSources].
	Type ObjectSourceType
	// Paths lists all resolved URIs the [ObjectSource] points at.
	Paths []string
	// Reader may be optionally provided with [ObjectSourceTypeReader]
	// for [ReadObjectsFromSources] to read from the [io.Reader].
	Reader io.Reader
	// Raw is the original, unresolved [RawObjectSource].
	// Example: a relative path which was resolved to its absolute form.
	Raw RawObjectSource
}

// String implements [fmt.Stringer] interface.
func (o ObjectSource) String() string {
	return fmt.Sprintf("{ObjectSourceType: %s, Raw: %s}", o.Type, o.Raw)
}

//go:generate ../bin/go-enum --names

// ObjectSourceType represents the source (where does it come from) of the [manifest.Object] definition.
/* ENUM(
File = 1
Directory
GlobPattern
URL
Reader
)*/
type ObjectSourceType int

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
func resolveFSPath(path string) (typ ObjectSourceType, paths []string, err error) {
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
		typ = ObjectSourceTypeDirectory
	default:
		typ = ObjectSourceTypeFile
		paths = []string{path}
	}
	return typ, paths, nil
}

// resolveGlobPattern evaluates patterns defined in [filepath.Match] documentation.
// It supports double '**' wildcards directory expansion via a 3rd party library: [doublestar].
// The library is a pure (no dependencies) implementation of glob patterns with double wildcard support.
// Whenever a double wildcard is detected doublestar.Glob is used, otherwise we
// use [filepath.Glob] to minimize the impact of potential bugs in [doublestar.Glob].
//
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
		// To keep it platform independent we need to make sure the RawObjectSource
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

// hasGlobMeta reports whether path contains any of the magic characters recognized by [filepath.Match].
// Copied from filepath/match.go.
func hasGlobMeta(path string) bool {
	magicChars := `*?[`
	if runtime.GOOS != "windows" {
		magicChars = `*?[\`
	}
	return strings.ContainsAny(path, magicChars)
}
