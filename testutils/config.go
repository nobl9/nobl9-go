package testutils

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nobl9/nobl9-go/sdk"
)

// Setup initializes the internal utilities used by this package, like [sdk.Client].
// It can be called any number of times, but only the first invocation has an effect.
func Setup(config Config) {
	setupOnce.Do(func() {
		client = config.Client
		toolName = config.ToolName
	})
}

// Config is used to configure tis package's behavior.
type Config struct {
	// ToolName
	ToolName string
	Client   *sdk.Client
}

var (
	client   *sdk.Client
	toolName string

	testStartTime             = time.Now()
	objectsCounter            = atomic.Int64{}
	uniqueTestIdentifierLabel = struct {
		Key   string
		Value string
	}{
		Key:   "sdk-e2e-test-id",
		Value: strconv.Itoa(int(testStartTime.UnixNano())),
	}
	applyAndDeleteLock = newApplyAndDeleteLocker()

	setupOnce sync.Once
)
