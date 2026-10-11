package display

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/makeworld-the-better-one/amfora/structs"
	"github.com/stretchr/testify/assert"
)

// TestTabMutexGuardedUnlock verifies that Unlock on an unlocked tab does not panic.
func TestTabMutexGuardedUnlock(t *testing.T) {
	tab := makeNewTab()

	// Calling Unlock on an unlocked tab must not panic
	assert.NotPanics(t, func() {
		tab.Unlock()
	})

	// Normal lock/unlock cycle
	assert.NotPanics(t, func() {
		tab.Lock()
		tab.Unlock()
	})

	// Redundant unlock must not panic
	assert.NotPanics(t, func() {
		tab.Unlock()
		tab.Unlock()
	})
}

// TestTabMutexGuardedRUnlock verifies that RUnlock on an unlocked tab does not panic.
func TestTabMutexGuardedRUnlock(t *testing.T) {
	tab := makeNewTab()

	// Calling RUnlock on an unlocked tab must not panic
	assert.NotPanics(t, func() {
		tab.RUnlock()
	})

	// Normal rlock/runlock cycle
	assert.NotPanics(t, func() {
		tab.RLock()
		tab.RUnlock()
	})

	// Redundant runlock must not panic
	assert.NotPanics(t, func() {
		tab.RUnlock()
		tab.RUnlock()
	})
}

// TestTabCleanupIdempotent verifies that Cleanup and Close can run multiple times across goroutines
// without causing mutex panics.
func TestTabCleanupIdempotent(t *testing.T) {
	tab := makeNewTab()

	// Initial cleanup
	assert.NotPanics(t, func() {
		tab.Cleanup()
	})
	assert.True(t, tab.IsClosed())

	// Secondary cleanup/close calls
	assert.NotPanics(t, func() {
		tab.Close()
		tab.Cleanup()
		tab.Unlock()
		tab.RUnlock()
	})

	// Concurrent cleanup calls
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tab.Close()
			tab.Cleanup()
			tab.Unlock()
			tab.RUnlock()
		}()
	}
	wg.Wait()
}

// TestTabConcurrentLockAndCleanup simulates background goroutines accessing/locking the tab
// while concurrent tab cleanup and close operations are performed.
func TestTabConcurrentLockAndCleanup(t *testing.T) {
	tab := makeNewTab()
	tab.page = &structs.Page{URL: "gemini://example.com", Content: "hello"}

	var wg sync.WaitGroup
	workers := 10
	iterations := 100

	// Readers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tab.RLock()
				_ = tab.hasContent()
				tab.RUnlock()
				time.Sleep(time.Microsecond * 10)
			}
		}()
	}

	// Writers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tab.Lock()
				if tab.page != nil {
					tab.page.SelectedID = strconv.Itoa(j)
				}
				tab.Unlock()
				time.Sleep(time.Microsecond * 10)
			}
		}()
	}

	// Concurrent closers/cleaners
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tab.Cleanup()
				tab.Unlock()
				tab.RUnlock()
				time.Sleep(time.Microsecond * 15)
			}
		}()
	}

	wg.Wait()
	assert.True(t, tab.IsClosed())
}

// TestConcurrentTabCloseAndReopen simulates concurrent tab/buffer closing and reopening
// verifying no fatal mutex panics or race conditions occur.
func TestConcurrentTabCloseAndReopen(t *testing.T) {
	// Reset tabs
	tabMu.Lock()
	tabs = []*tab{makeNewTab()}
	curTab = 0
	tabMu.Unlock()

	var wg sync.WaitGroup
	numOps := 50

	// Goroutine opening tabs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			tabMu.Lock()
			t := makeNewTab()
			t.page = &structs.Page{URL: "about:newtab"}
			tabs = append(tabs, t)
			curTab = len(tabs) - 1
			tabMu.Unlock()
			time.Sleep(time.Millisecond)
		}
	}()

	// Goroutine closing tabs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numOps; i++ {
			tabMu.Lock()
			if len(tabs) > 1 {
				closing := tabs[len(tabs)-1]
				tabs = tabs[:len(tabs)-1]
				closing.Close()
				if curTab >= len(tabs) {
					curTab = len(tabs) - 1
				}
			}
			tabMu.Unlock()
			time.Sleep(time.Millisecond)
		}
	}()

	// Goroutines inspecting/accessing tabs
	for w := 0; w < 3; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < numOps; i++ {
				n := NumTabs()
				if n > 0 {
					tabMu.RLock()
					if curTab >= 0 && curTab < len(tabs) {
						current := tabs[curTab]
						if current != nil && !current.IsClosed() {
							current.RLock()
							_ = current.hasContent()
							current.RUnlock()
						}
					}
					tabMu.RUnlock()
				}
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Clean up after test
	tabMu.Lock()
	for _, t := range tabs {
		t.Close()
	}
	tabs = nil
	curTab = -1
	tabMu.Unlock()
}
