package e2etestutils

import (
	"strconv"
	"strings"
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

// Config is used to configure this package's behavior.
type Config struct {
	// ToolName is the name of the tested tool, e.g. 'SDK', 'Terraform', 'sloctl'.
	ToolName string
	// Client is used to shared by all functions which interact with the Nobl9 API.
	Client *sdk.Client
}

var (
	client   *sdk.Client
	toolName string

	testStartTime          = time.Now()
	objectsCounter         = atomic.Int64{}
	uniqueTestIDLabelValue = strconv.Itoa(int(testStartTime.UnixNano()))
	applyAndDeleteLock     = newApplyAndDeleteLocker()

	setupOnce sync.Once
)

func getUniqueTestIDLabelKey() string {
	return strings.ToLower(toolName) + "-e2e-test-id"
}
