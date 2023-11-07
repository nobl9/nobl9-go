package testutils

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/rs/zerolog/log"
)

var (
	rec              = testRecorder{}
	initRecorderOnce sync.Once
)

type testRecorder struct {
	shouldRecord bool
	mu           sync.Mutex
	output       io.Writer
}

type recordedTest struct {
	TestName    string          `json:"testName"`
	Object      interface{}     `json:"object"`
	ErrorsCount int             `json:"errorsCount"`
	Errors      []ExpectedError `json:"errors"`
}

func (r *testRecorder) Record(t *testing.T, object interface{}, errorsCount int, errors []ExpectedError) {
	r.Init()
	if !r.shouldRecord {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := json.NewEncoder(r.output).Encode(recordedTest{
		TestName:    t.Name(),
		Object:      object,
		ErrorsCount: errorsCount,
		Errors:      errors,
	}); err != nil {
		log.Err(err).Msg("failed to record test")
	}
}

func (r *testRecorder) Init() {
	initRecorderOnce.Do(func() {
		path, isSet := os.LookupEnv("NOBL9_SDK_TEST_RECORD_FILE")
		if !isSet {
			return
		}
		r.shouldRecord = true
		// #nosec G304
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			panic(err)
		}
		r.output = f
		log.Info().Msg("test recorder initialized, all test will be recorded in " + path)
	})
}
