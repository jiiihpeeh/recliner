package hooks

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jiiihpeeh/recliner/events"
	stores "github.com/jiiihpeeh/recliner/store"
)

var (
	globalHooksCtx *HooksContext
	hooksCtxOnce   sync.Once
	hooksLogFile   *os.File
	debugCallback  func(string, string) // Callback for debug messages (category, message)
)

func SetDebugCallback(callback func(string, string)) {
	debugCallback = callback
}

func logHookEntry(entry string) {
	if hooksLogFile == nil {
		return
	}
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(hooksLogFile, "[%s] %s\n", timestamp, entry)
	hooksLogFile.Sync()

	// Send to debug callback if set
	if debugCallback != nil {
		debugCallback("HOOKS", fmt.Sprintf("%s", entry))
	}
}

func logHookChange(index int, oldVal any, newVal any) {
	logHookEntry(fmt.Sprintf("hook[%d]: %v -> %v", index, oldVal, newVal))
}

func LogKeyPress(event events.KeyPressEvent) {
	logHookEntry(fmt.Sprintf(
		"key: %s escape=%v ctrl=%v shift=%v meta=%v",
		event.Key,
		event.Escape,
		event.Ctrl,
		event.Shift,
		event.Meta,
	))
}

type State struct {
	Value any
}

type Effect struct {
	Func         func()
	Dependencies []any
	Done         bool
	Cleanup      func()
}

type Memo struct {
	Value        any
	Dependencies []any
}

type Ref[T any] struct {
	Value T
}

// Event types live in the `events` package. Use `events.*` directly.

type HooksContext struct {
	mu                sync.RWMutex
	states            []State
	effects           []Effect
	layoutEffects     []Effect
	memos             []Memo
	refs              []any
	stateIndex        int
	effectIndex       int
	layoutEffectIndex int
	memoIndex         int
	refIndex          int
	inputIndex        int
	activeInputs      int
	InputHandlers     []func(events.KeyPressEvent)
	onUpdate          func()
	stores            []storeSubscription
	storeIndex        int

	// App Context
	App    any
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File

	// Focus Management
	FocusID      string
	FocusableIDs []string

	// Cursor Management
	CursorX       int
	CursorY       int
	CursorVisible bool

	pendingEffects       []func()
	pendingEffectsQueue  []func() // Effects discovered during render
	pendingLayoutEffects []func()

	WindowWidth  int
	WindowHeight int
}

type storeSubscription struct {
	unsub func()
}

func NewHooksContext() *HooksContext {
	return &HooksContext{
		states:            make([]State, 0),
		effects:           make([]Effect, 0),
		layoutEffects:     make([]Effect, 0),
		memos:             make([]Memo, 0),
		refs:              make([]any, 0),
		stateIndex:        0,
		effectIndex:       0,
		layoutEffectIndex: 0,
		memoIndex:         0,
		refIndex:          0,
		InputHandlers:     make([]func(events.KeyPressEvent), 0),
		stores:            make([]storeSubscription, 0),
		WindowWidth:       80,
		WindowHeight:      24,
	}
}

func (hc *HooksContext) SetWindowSize(w, h int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	if hc.WindowWidth != w || hc.WindowHeight != h {
		hc.WindowWidth = w
		hc.WindowHeight = h
	}
}

func UseWindowSize(hc *HooksContext) (int, int) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.WindowWidth, hc.WindowHeight
}

func GetContext() *HooksContext {
	hooksCtxOnce.Do(func() {
		if globalHooksCtx == nil {
			globalHooksCtx = NewHooksContext()
		}
	})
	if globalHooksCtx == nil {
		globalHooksCtx = NewHooksContext()
	}
	return globalHooksCtx
}

func SetContext(ctx *HooksContext) {
	globalHooksCtx = ctx
}

func (hc *HooksContext) Reset() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.stateIndex = 0
	hc.effectIndex = 0
	hc.layoutEffectIndex = 0
	hc.memoIndex = 0
	hc.refIndex = 0
	hc.inputIndex = 0
	hc.storeIndex = 0

	// Reset focusable IDs for new render pass
	hc.FocusableIDs = hc.FocusableIDs[:0]

	// Reset cursor state for new render pass (default hidden)
	hc.CursorVisible = false
	hc.CursorX = 0
	hc.CursorY = 0

	// Clear pending queues from previous aborted renders if any
	hc.pendingEffectsQueue = hc.pendingEffectsQueue[:0]
	hc.pendingLayoutEffects = hc.pendingLayoutEffects[:0]
}

func (hc *HooksContext) Finalize() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.activeInputs = hc.inputIndex

	// Queue non-layout effects from this pass
	hc.pendingEffects = append(hc.pendingEffects, hc.pendingEffectsQueue...)
	hc.pendingEffectsQueue = hc.pendingEffectsQueue[:0]

	// Prune unused state/hooks to prevent slow accumulation
	if hc.stateIndex < len(hc.states) {
		hc.states = hc.states[:hc.stateIndex]
	}
	if hc.effectIndex < len(hc.effects) {
		// Run cleanup for pruned effects
		for i := hc.effectIndex; i < len(hc.effects); i++ {
			if hc.effects[i].Cleanup != nil {
				hc.effects[i].Cleanup()
			}
		}
		hc.effects = hc.effects[:hc.effectIndex]
	}
	if hc.layoutEffectIndex < len(hc.layoutEffects) {
		// Run cleanup for pruned layout effects
		for i := hc.layoutEffectIndex; i < len(hc.layoutEffects); i++ {
			if hc.layoutEffects[i].Cleanup != nil {
				hc.layoutEffects[i].Cleanup()
			}
		}
		hc.layoutEffects = hc.layoutEffects[:hc.layoutEffectIndex]
	}
	if hc.memoIndex < len(hc.memos) {
		hc.memos = hc.memos[:hc.memoIndex]
	}
	if hc.refIndex < len(hc.refs) {
		hc.refs = hc.refs[:hc.refIndex]
	}
	if hc.storeIndex < len(hc.stores) {
		// Unsubscribe from pruned stores
		for i := hc.storeIndex; i < len(hc.stores); i++ {
			if hc.stores[i].unsub != nil {
				hc.stores[i].unsub()
			}
		}
		hc.stores = hc.stores[:hc.storeIndex]
	}
}

func (hc *HooksContext) GetInputHandlers() []func(events.KeyPressEvent) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	count := min(hc.activeInputs, len(hc.InputHandlers))
	handlers := make([]func(events.KeyPressEvent), count)
	copy(handlers, hc.InputHandlers[:count])
	return handlers
}

func (hc *HooksContext) triggerUpdate() {
	hc.mu.RLock()
	cb := hc.onUpdate
	hc.mu.RUnlock()

	if cb != nil {
		cb()
	} else if globalHooksCtx != nil && globalHooksCtx != hc {
		globalHooksCtx.triggerUpdate()
	}
}

func (hc *HooksContext) SetUpdateCallback(fn func()) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.onUpdate = fn
}

func (hc *HooksContext) FlushLayoutEffects() {
	hc.mu.Lock()
	tasks := make([]func(), len(hc.pendingLayoutEffects))
	copy(tasks, hc.pendingLayoutEffects)
	hc.pendingLayoutEffects = hc.pendingLayoutEffects[:0]
	hc.mu.Unlock()

	for _, task := range tasks {
		task()
	}
}

func (hc *HooksContext) FlushEffects() {
	hc.mu.Lock()
	tasks := make([]func(), len(hc.pendingEffects))
	copy(tasks, hc.pendingEffects)
	hc.pendingEffects = hc.pendingEffects[:0]
	hc.mu.Unlock()

	for _, task := range tasks {
		task()
	}
}

func UseState[T any](hc *HooksContext, initialValue T) (T, func(T)) {
	hc.mu.Lock()
	idx := hc.stateIndex
	if idx >= len(hc.states) {
		hc.states = append(hc.states, State{Value: initialValue})
	}
	value := hc.states[idx].Value
	hc.stateIndex++
	hc.mu.Unlock()

	var currentValue T
	if value != nil {
		if v, ok := value.(T); ok {
			currentValue = v
		} else {
			currentValue = initialValue
		}
	} else {
		currentValue = initialValue
	}

	setter := func(newValue T) {
		hc.mu.Lock()
		if idx >= 0 && idx < len(hc.states) {
			hc.states[idx].Value = newValue
		}
		hc.mu.Unlock()

		hc.triggerUpdate()
	}

	return currentValue, setter
}

func UseReducer[S any, A any](hc *HooksContext, reducer func(S, A) S, initialState S) (S, func(A)) {
	hc.mu.Lock()
	idx := hc.stateIndex
	if idx >= len(hc.states) {
		hc.states = append(hc.states, State{Value: initialState})
	}
	value := hc.states[idx].Value
	hc.stateIndex++
	hc.mu.Unlock()

	currentState, ok := value.(S)
	if !ok {
		currentState = initialState
	}

	dispatch := func(action A) {
		hc.mu.Lock()
		var oldVal any
		var newState S
		if idx >= 0 && idx < len(hc.states) {
			oldVal = hc.states[idx].Value

			oldS, ok := oldVal.(S)
			if !ok {
				oldS = initialState
			}

			newState = reducer(oldS, action)
			hc.states[idx].Value = newState
		}
		hc.mu.Unlock()

		hc.triggerUpdate()
	}

	return currentState, dispatch
}

func UseEffect(hc *HooksContext, effect func() func(), deps []any) {
	hc.UseEffect(effect, deps)
}

func (hc *HooksContext) UseEffect(effect func() func(), deps []any) {
	hc.mu.Lock()
	idx := hc.effectIndex
	if idx >= len(hc.effects) {
		hc.effects = append(hc.effects, Effect{})
	}
	current := hc.effects[idx]
	hc.effectIndex++
	hc.mu.Unlock()

	shouldRun := !current.Done || len(current.Dependencies) != len(deps)
	if !shouldRun {
		for i, dep := range deps {
			if !equalAny(dep, current.Dependencies[i]) {
				shouldRun = true
				break
			}
		}
	}

	if shouldRun {
		// Queue discovery
		effectIdx := idx // Capture for closure
		hc.mu.Lock()
		hc.effects[idx].Dependencies = deps
		hc.effects[idx].Done = true

		hc.pendingEffectsQueue = append(hc.pendingEffectsQueue, func() {
			hc.mu.Lock()
			// Re-read cleanup from storage as it might have been updated
			cleanup := hc.effects[effectIdx].Cleanup
			hc.mu.Unlock()

			if cleanup != nil {
				cleanup()
			}
			newCleanup := effect()

			hc.mu.Lock()
			hc.effects[effectIdx].Cleanup = newCleanup
			hc.mu.Unlock()
		})
		hc.mu.Unlock()
	}
}

func (hc *HooksContext) UseLayoutEffect(effect func() func(), deps []any) {
	hc.mu.Lock()
	idx := hc.layoutEffectIndex
	if idx >= len(hc.layoutEffects) {
		hc.layoutEffects = append(hc.layoutEffects, Effect{})
	}
	current := hc.layoutEffects[idx]
	hc.layoutEffectIndex++
	hc.mu.Unlock()

	shouldRun := !current.Done || len(current.Dependencies) != len(deps)
	if !shouldRun {
		for i, dep := range deps {
			if !equalAny(dep, current.Dependencies[i]) {
				shouldRun = true
				break
			}
		}
	}

	if shouldRun {
		hc.mu.Lock()
		hc.layoutEffects[idx].Dependencies = deps
		hc.layoutEffects[idx].Done = true

		effectIdx := idx // Capture for closure
		hc.pendingLayoutEffects = append(hc.pendingLayoutEffects, func() {
			hc.mu.Lock()
			cleanup := hc.layoutEffects[effectIdx].Cleanup
			hc.mu.Unlock()

			if cleanup != nil {
				cleanup()
			}
			newCleanup := effect()

			hc.mu.Lock()
			hc.layoutEffects[effectIdx].Cleanup = newCleanup
			hc.mu.Unlock()
		})
		hc.mu.Unlock()
	}
}

func (hc *HooksContext) UseInput(handler func(events.KeyPressEvent), deps []any) {
	_ = deps
	hc.mu.Lock()
	idx := hc.inputIndex
	if idx >= len(hc.InputHandlers) {
		hc.InputHandlers = append(hc.InputHandlers, handler)
	} else {
		hc.InputHandlers[idx] = handler
	}
	hc.inputIndex++
	hc.mu.Unlock()
}

func (hc *HooksContext) Cleanup() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for i := range hc.effects {
		if hc.effects[i].Cleanup != nil {
			hc.effects[i].Cleanup()
			hc.effects[i].Cleanup = nil
		}
	}

	for i := range hc.layoutEffects {
		if hc.layoutEffects[i].Cleanup != nil {
			hc.layoutEffects[i].Cleanup()
			hc.layoutEffects[i].Cleanup = nil
		}
	}

	hc.InputHandlers = nil
	for _, sub := range hc.stores {
		if sub.unsub != nil {
			sub.unsub()
		}
	}
	hc.stores = nil
}

func equalAny(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func UseMemo[T any](hc *HooksContext, factory func() T, deps []any) T {
	hc.mu.Lock()
	idx := hc.memoIndex
	if idx >= len(hc.memos) {
		hc.memos = append(hc.memos, Memo{})
	}
	current := hc.memos[idx]
	hc.memoIndex++
	hc.mu.Unlock()

	shouldCompute := false
	if deps == nil {
		shouldCompute = true
	} else if current.Dependencies == nil {
		shouldCompute = true
	} else if len(current.Dependencies) != len(deps) {
		shouldCompute = true
	} else {
		for i, dep := range deps {
			if !equalAny(dep, current.Dependencies[i]) {
				shouldCompute = true
				break
			}
		}
	}

	if shouldCompute {
		newValue := factory()
		hc.mu.Lock()
		hc.memos[idx] = Memo{
			Value:        newValue,
			Dependencies: deps,
		}
		hc.mu.Unlock()
		return newValue
	}

	if current.Value == nil {
		var zero T
		return zero
	}

	if v, ok := current.Value.(T); ok {
		return v
	}

	var zero T
	return zero
}

func UseCallback[T any](hc *HooksContext, fn T, deps []any) T {
	return UseMemo(hc, func() T {
		return fn
	}, deps)
}

func UseStore[T any](hc *HooksContext, s *stores.Store[T]) T {
	if s == nil {
		var zero T
		return zero
	}
	state, setState := UseState(hc, s.Get())

	hc.mu.Lock()
	idx := hc.storeIndex
	hc.storeIndex++
	if idx >= len(hc.stores) {
		hc.stores = append(hc.stores, storeSubscription{})
	}
	sub := hc.stores[idx]
	hc.mu.Unlock()

	if sub.unsub == nil {
		unsub := s.Subscribe(func(val T) {
			setState(val)
		})
		hc.mu.Lock()
		if idx < len(hc.stores) {
			hc.stores[idx].unsub = unsub
		}
		hc.mu.Unlock()
	}

	return state
}

func UseRef[T any](hc *HooksContext, initialValue T) *Ref[T] {
	hc.mu.Lock()
	idx := hc.refIndex
	var ref *Ref[T]
	if idx < len(hc.refs) {
		if existing, ok := hc.refs[idx].(*Ref[T]); ok && existing != nil {
			ref = existing
		}
	}
	if ref == nil {
		ref = &Ref[T]{Value: initialValue}
		if idx >= len(hc.refs) {
			hc.refs = append(hc.refs, ref)
		} else {
			hc.refs[idx] = ref
		}
	}
	hc.refIndex++
	hc.mu.Unlock()
	return ref
}
