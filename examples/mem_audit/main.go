package main

import (
	"fmt"
	"runtime"

	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/c"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/store"
	"github.com/jiiihpeeh/recliner/vdom"
)

var testStore = store.NewStore(0)

func main() {
	fmt.Println("Starting Memory Leak Audit Test...")

	// Track initial memory
	var ms runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&ms)
	initialAlloc := ms.Alloc
	fmt.Printf("Initial Alloc: %d KB\n", initialAlloc/1024)

	// Create an app instance with a root component that triggers frequent re-renders
	// and uses various hooks to see if they accumulate.
	appInstance := app.NewWithOptions(LeakTestRoot, app.AppOptions{
		Headless: true,
	})

	// Run for a fixed number of iterations
	iterations := 10000
	for i := 0; i < iterations; i++ {
		_, err := appInstance.RunHeadless()
		if err != nil {
			panic(err)
		}

		if i%100 == 0 {
			runtime.ReadMemStats(&ms)
			fmt.Printf("Iteration %d: Alloc = %d KB\n", i, ms.Alloc/1024)
		}
	}

	// Final report
	runtime.GC()
	runtime.ReadMemStats(&ms)
	finalAlloc := ms.Alloc
	fmt.Printf("Final Alloc after %d iterations: %d KB\n", iterations, finalAlloc/1024)
	fmt.Printf("Diff: %d KB\n", (finalAlloc-initialAlloc)/1024)

	if finalAlloc > initialAlloc+(1024*1024) { // 1MB threshold for simple test
		fmt.Println("POTENTIAL MEMORY LEAK DETECTED")
	} else {
		fmt.Println("No significant memory leak detected in basic test.")
	}
}

func LeakTestRoot(props any) vdom.Node {
	hc := hooks.GetContext()

	// UseState
	count, setCount := hooks.UseState(hc, 0)

	// UseEffect with cleanup
	hooks.GetContext().UseEffect(func() func() {
		// Simulation of a subscription
		return func() {
			// Cleanup logic
		}
	}, []any{count})

	// UseMemo
	_ = hooks.UseMemo(hc, func() string {
		return fmt.Sprintf("Computed %d", count)
	}, []any{count})

	// UseStore
	_ = hooks.UseStore(hc, testStore)

	// Leak: UseEffect that never runs cleanup because deps are always changing but not calling it?
	// Actually UseEffect DOES call cleanup if it runs again.
	// What if it never runs again?
	if count == 0 {
		hooks.GetContext().UseEffect(func() func() {
			return func() {
				fmt.Println("This should only run once at the very end if ever")
			}
		}, []any{})
	}

	setCount(count + 1)

	return c.Box(c.BoxProps{},
		c.Text(c.TextProps{Content: fmt.Sprintf("Iteration: %d", count)}),
	)
}
