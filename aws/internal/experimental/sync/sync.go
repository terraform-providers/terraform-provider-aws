package sync

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
)

// Semaphore can be used to limit concurrent executions. This can be used to work with resources with low quotas
type Semaphore chan struct{}

// InitializeSemaphore initializes a semaphore with a default capacity or overrides it using an environment variable
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func InitializeSemaphore(envvar string, defaultLimit int) Semaphore {
	limit := defaultLimit
	x := os.Getenv(envvar)
	if x != "" {
		var err error
		limit, err = strconv.Atoi(x)
		if err != nil {
			panic(fmt.Errorf("could not parse %q: expected integer, got %q", envvar, x))
		}
	}
	return make(Semaphore, limit)
}

// Wait waits for a semaphore before continuing
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func (s Semaphore) Wait() {
	s <- struct{}{}
}

// Notify releases a semaphore
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func (s Semaphore) Notify() {
	// Make the Notify non-blocking. This can happen if a Wait was never issued
	select {
	case <-s:
	default:
		log.Println("[WARN] Notifying semaphore without Wait")
	}
}

// TestAccPreCheckSyncronized waits for a semaphore and skips the test if there is no capacity
// NOTE: this is currently an experimental feature and is likely to change. DO NOT USE.
func TestAccPreCheckSyncronize(t *testing.T, semaphore Semaphore, resource string) {
	if cap(semaphore) == 0 {
		t.Skipf("concurrency for %s testing set to 0", resource)
	}

	semaphore.Wait()
}
