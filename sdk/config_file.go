package sdk

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// FileConfig contains fully parsed config file.
type FileConfig struct {
	ContextlessConfig `toml:",inline"`
	Contexts          map[string]ContextConfig `toml:"contexts" json:"contexts"`

	filePath string
}

// GetCurrentContextConfig retrieves the [ContextConfig] for the current default context.
func (f *FileConfig) GetCurrentContextConfig() (ContextConfig, error) {
	return f.GetContextConfig(f.DefaultContext)
}

// GetContextConfig retrieves the [ContextConfig] associated with the provided context name.
func (f *FileConfig) GetContextConfig(contextName string) (ContextConfig, error) {
	config, ok := f.Contexts[contextName]
	if !ok {
		return config, errors.Errorf("no config defined for %q context", contextName)
	}
	return config, nil
}

// GetPath retrieves the file path [FileConfig] was loaded from.
func (f *FileConfig) GetPath() string {
	return f.filePath
}

// Load reads the config file from the provided path.
// If the file does not exist, it will create a default configuration file at the provided path.
func (f *FileConfig) Load(path string) error {
	f.filePath = path
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to stat config file: %s", path)
		}
		if err = createDefaultConfigFile(path); err != nil {
			return err
		}
	}
	if _, err := toml.DecodeFile(path, &f); err != nil {
		return errors.Wrapf(err, "could not decode config file: %s", path)
	}
	return nil
}

// Save saves [FileConfig] into provided path, encoding it in TOML format.
func (f *FileConfig) Save(path string) error {
	tmpFileName, err := f.writeToTempFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create and write a temporary config file used for saving the config changes")
	}
	if err = os.Rename(tmpFileName, path); err != nil {
		return err
	}
	f.filePath = path
	return nil
}

func (f *FileConfig) writeToTempFile(path string) (tmpFileName string, err error) {
	tmpFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return "", err
	}
	defer func() { _ = tmpFile.Close() }()
	if err = toml.NewEncoder(tmpFile).Encode(f); err != nil {
		return "", err
	}
	if err = tmpFile.Sync(); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func createDefaultConfigFile(path string) error {
	dir := filepath.Dir(path)
	// Create the directory with all it's parents.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0o700); err != nil {
			return errors.Wrapf(err, "failed to create a directory path (with parents) for %s", dir)
		}
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat %s directory", dir)
	}
	// Create the config file.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// #nosec G304
		f, err := os.Create(path)
		if err != nil {
			return errors.Wrapf(err, "failed to create Nobl9 config file under %s", path)
		}
		defer func() { _ = f.Close() }()
		return toml.NewEncoder(f).Encode(FileConfig{
			ContextlessConfig: ContextlessConfig{DefaultContext: defaultContext},
			Contexts:          map[string]ContextConfig{defaultContext: {}},
		})
	} else if err != nil {
		return errors.Wrapf(err, "failed to stat %s file", path)
	}
	return nil
}
