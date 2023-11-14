package testutils

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/rs/zerolog/log"
)

var rec = new(testRecorder)

type testRecorder struct {
	shouldRecord bool
	output       io.Writer
	mu           sync.Mutex
	initOnce     sync.Once
}

type recordedTest struct {
	TestName    string          `json:"testName"`
	Object      interface{}     `json:"object"`
	IsValid     bool            `json:"isValid"`
	ErrorsCount int             `json:"errorsCount,omitempty"`
	Errors      []ExpectedError `json:"errors,omitempty"`
}

func (r *testRecorder) Record(t *testing.T, object interface{}, errorsCount int, errors []ExpectedError) {
	r.Init()
	if !r.shouldRecord {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	rt := recordedTest{
		TestName:    t.Name(),
		Object:      object,
		ErrorsCount: errorsCount,
		Errors:      errors,
	}
	if errorsCount == 0 {
		rt.IsValid = true
	}
	if err := json.NewEncoder(r.output).Encode(rt); err != nil {
		log.Err(err).Msg("failed to record test")
	}
}

func (r *testRecorder) Init() {
	r.initOnce.Do(func() {
		path, isSet := os.LookupEnv("NOBL9_SDK_TEST_RECORD_FILE")
		if !isSet {
			return
		}
		r.shouldRecord = true
		// #nosec G304
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		r.output = f
		log.Info().Msg("test recorder initialized, all test will be recorded in " + path)
	})
}
