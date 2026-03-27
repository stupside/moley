package framework_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stupside/moley/v2/internal/platform/framework"
)

// --- Test Lifecycle implementation ---

type testInput struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type testOutput struct {
	Name    string `json:"name"`
	Created bool   `json:"created"`
}

type testHandler struct {
	handlerName string
	created     map[string]testOutput
	destroyed   map[string]bool
}

func newTestHandler(name string) *testHandler {
	return &testHandler{
		handlerName: name,
		created:     make(map[string]testOutput),
		destroyed:   make(map[string]bool),
	}
}

func (h *testHandler) Name() string           { return h.handlerName }
func (h *testHandler) Key(i testInput) string { return i.Name }

func (h *testHandler) Create(_ context.Context, input testInput) (testOutput, error) {
	out := testOutput{Name: input.Name, Created: true}
	h.created[input.Name] = out
	return out, nil
}

func (h *testHandler) Destroy(_ context.Context, output testOutput) error {
	h.destroyed[output.Name] = true
	return nil
}

func (h *testHandler) Check(_ context.Context, output testOutput) (framework.Status, error) {
	if output.Created {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}

func (h *testHandler) Recover(_ context.Context, _ testInput) (testOutput, framework.Status, error) {
	return testOutput{}, framework.StatusDown, nil
}

// --- Helpers ---

func chdir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	orig, _ := os.Getwd()
	os.Chdir(dir)
	t.Cleanup(func() { os.Chdir(orig) })
}

func hashJSON(v any) string {
	data, _ := json.Marshal(v)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func staticResolver(key string) framework.InputResolver[testInput] {
	return func(_ *framework.OutputRegistry) ([]testInput, error) {
		return []testInput{{Name: key, Value: 0}}, nil
	}
}

// --- Hash computation tests ---

func TestHashDeterministic(t *testing.T) {
	h1 := hashJSON(testInput{Name: "foo", Value: 42})
	h2 := hashJSON(testInput{Name: "foo", Value: 42})
	if h1 != h2 {
		t.Errorf("same input produced different hashes: %s vs %s", h1, h2)
	}
}

func TestHashChangesOnDifferentInput(t *testing.T) {
	h1 := hashJSON(testInput{Name: "foo", Value: 1})
	h2 := hashJSON(testInput{Name: "foo", Value: 2})
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}

// --- Lock file tests ---

func TestLockFileSaveAndLoad(t *testing.T) {
	chdir(t)

	lf, err := framework.LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}

	lf.Entries = append(lf.Entries, framework.LockEntry{
		Key:         "test-key",
		HandlerName: "test-handler",
		InputHash:   "abc123",
		Data:        map[string]any{"foo": "bar"},
	})

	if err := lf.Save(); err != nil {
		t.Fatal(err)
	}
	_ = lf.Close()

	// Reload
	lf2, err := framework.LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lf2.Close() }()

	if len(lf2.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(lf2.Entries))
	}
	if lf2.Entries[0].Key != "test-key" {
		t.Errorf("expected key 'test-key', got %q", lf2.Entries[0].Key)
	}
	if lf2.Entries[0].InputHash != "abc123" {
		t.Errorf("expected hash 'abc123', got %q", lf2.Entries[0].InputHash)
	}
}

func TestLockFileCorruptRecovery(t *testing.T) {
	chdir(t)
	os.WriteFile("moley.lock", []byte("not json{{{"), 0644)

	lf, err := framework.LoadLockFile()
	if err != nil {
		t.Fatalf("corrupt lock file should not error: %v", err)
	}
	defer func() { _ = lf.Close() }()

	if len(lf.Entries) != 0 {
		t.Errorf("corrupt lock file should load as empty, got %d entries", len(lf.Entries))
	}
}

func TestLockFileEmptyRecovery(t *testing.T) {
	chdir(t)
	os.WriteFile("moley.lock", []byte(""), 0644)

	lf, err := framework.LoadLockFile()
	if err != nil {
		t.Fatalf("empty lock file should not error: %v", err)
	}
	defer func() { _ = lf.Close() }()

	if len(lf.Entries) != 0 {
		t.Errorf("empty lock file should load as empty, got %d entries", len(lf.Entries))
	}
}

func TestLockFileMissingIsEmpty(t *testing.T) {
	chdir(t)

	lf, err := framework.LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lf.Close() }()

	if len(lf.Entries) != 0 {
		t.Errorf("missing lock file should load as empty, got %d entries", len(lf.Entries))
	}
}

func TestLockFilePurgeOrphans(t *testing.T) {
	chdir(t)

	lf, err := framework.LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lf.Close() }()

	lf.Entries = []framework.LockEntry{
		{Key: "a", HandlerName: "keep-1"},
		{Key: "b", HandlerName: "orphan"},
		{Key: "c", HandlerName: "keep-2"},
	}

	if err := lf.PurgeOrphans(map[string]bool{"keep-1": true, "keep-2": true}); err != nil {
		t.Fatal(err)
	}

	if len(lf.Entries) != 2 {
		t.Fatalf("expected 2 entries after purge, got %d", len(lf.Entries))
	}
	for _, e := range lf.Entries {
		if e.HandlerName == "orphan" {
			t.Error("orphaned entry should have been purged")
		}
	}
}

func TestLockFileSequentialLocking(t *testing.T) {
	chdir(t)

	lf1, err := framework.LoadLockFile()
	if err != nil {
		t.Fatal(err)
	}
	_ = lf1.Close()

	// Second lock after release should succeed
	lf2, err := framework.LoadLockFile()
	if err != nil {
		t.Fatalf("second lock after release should succeed: %v", err)
	}
	_ = lf2.Close()
}

// --- Topological sort tests ---

type orderTracker struct {
	name         string
	order        *[]string
	destroyOrder *[]string
}

func (h *orderTracker) Name() string           { return h.name }
func (h *orderTracker) Key(_ testInput) string { return "single" }

func (h *orderTracker) Create(_ context.Context, input testInput) (testOutput, error) {
	*h.order = append(*h.order, h.name)
	return testOutput{Name: input.Name, Created: true}, nil
}

func (h *orderTracker) Destroy(_ context.Context, _ testOutput) error {
	if h.destroyOrder != nil {
		*h.destroyOrder = append(*h.destroyOrder, h.name)
	}
	return nil
}

func (h *orderTracker) Check(_ context.Context, output testOutput) (framework.Status, error) {
	if output.Created {
		return framework.StatusUp, nil
	}
	return framework.StatusDown, nil
}

func (h *orderTracker) Recover(_ context.Context, _ testInput) (testOutput, framework.Status, error) {
	return testOutput{}, framework.StatusDown, nil
}

func TestTopoSortLinearChain(t *testing.T) {
	chdir(t)

	var order []string
	h1 := &orderTracker{name: "step-1", order: &order}
	h2 := &orderTracker{name: "step-2", order: &order}
	h3 := &orderTracker{name: "step-3", order: &order}

	r, _ := framework.NewReconciler()
	framework.Register(r, h1, staticResolver("a"))
	framework.Register(r, h2, staticResolver("a"), "step-1")
	framework.Register(r, h3, staticResolver("a"), "step-2")

	if err := r.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 creates, got %d: %v", len(order), order)
	}
	if order[0] != "step-1" || order[1] != "step-2" || order[2] != "step-3" {
		t.Errorf("wrong order: %v", order)
	}
}

func TestTopoSortDiamond(t *testing.T) {
	chdir(t)

	var order []string
	r, _ := framework.NewReconciler()
	framework.Register(r, &orderTracker{name: "root", order: &order}, staticResolver("a"))
	framework.Register(r, &orderTracker{name: "left", order: &order}, staticResolver("a"), "root")
	framework.Register(r, &orderTracker{name: "right", order: &order}, staticResolver("a"), "root")
	framework.Register(r, &orderTracker{name: "bottom", order: &order}, staticResolver("a"), "left", "right")

	if err := r.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	if order[0] != "root" {
		t.Errorf("root should be first, got %v", order)
	}
	if order[3] != "bottom" {
		t.Errorf("bottom should be last, got %v", order)
	}
}

func TestTopoSortCircularDependency(t *testing.T) {
	chdir(t)

	r, _ := framework.NewReconciler()
	framework.Register(r, &orderTracker{name: "a", order: new([]string)}, staticResolver("x"), "b")
	framework.Register(r, &orderTracker{name: "b", order: new([]string)}, staticResolver("x"), "a")

	err := r.Start(context.Background())
	if err == nil {
		t.Fatal("circular dependency should cause an error")
	}
}

func TestTopoSortUnknownDependency(t *testing.T) {
	chdir(t)

	r, _ := framework.NewReconciler()
	framework.Register(r, &orderTracker{name: "a", order: new([]string)}, staticResolver("x"), "nonexistent")

	err := r.Start(context.Background())
	if err == nil {
		t.Fatal("unknown dependency should cause an error")
	}
}

func TestStopReverseOrder(t *testing.T) {
	chdir(t)
	ctx := context.Background()

	var startOrder, stopOrder []string

	h1 := &orderTracker{name: "first", order: &startOrder, destroyOrder: &stopOrder}
	h2 := &orderTracker{name: "second", order: &startOrder, destroyOrder: &stopOrder}
	h3 := &orderTracker{name: "third", order: &startOrder, destroyOrder: &stopOrder}

	r, _ := framework.NewReconciler()
	framework.Register(r, h1, staticResolver("a"))
	framework.Register(r, h2, staticResolver("a"), "first")
	framework.Register(r, h3, staticResolver("a"), "second")

	if err := r.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// New reconciler for stop (matches real usage)
	r2, _ := framework.NewReconciler()
	framework.Register(r2, h1, staticResolver("a"))
	framework.Register(r2, h2, staticResolver("a"), "first")
	framework.Register(r2, h3, staticResolver("a"), "second")

	if err := r2.Stop(ctx); err != nil {
		t.Fatal(err)
	}

	if len(stopOrder) != 3 {
		t.Fatalf("expected 3 destroys, got %d: %v", len(stopOrder), stopOrder)
	}
	if stopOrder[0] != "third" || stopOrder[1] != "second" || stopOrder[2] != "first" {
		t.Errorf("stop should be reverse order: %v", stopOrder)
	}
}

// --- Output registry / data flow tests ---

func TestOutputFlowsBetweenNodes(t *testing.T) {
	chdir(t)

	var receivedName string

	upstream := newTestHandler("upstream")

	r, _ := framework.NewReconciler()
	framework.Register(r, upstream,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "item", Value: 42}}, nil
		},
	)
	framework.Register(r, newTestHandler("downstream"),
		func(reg *framework.OutputRegistry) ([]testInput, error) {
			out, ok := framework.GetOutput[testOutput](reg, "upstream", "item")
			if !ok {
				return nil, fmt.Errorf("missing upstream output")
			}
			receivedName = out.Name
			return []testInput{{Name: "derived"}}, nil
		},
		"upstream",
	)

	if err := r.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	if receivedName != "item" {
		t.Errorf("downstream should have received 'item', got %q", receivedName)
	}
}

// --- Hash change detection tests ---

func TestHashChangeTriggersUpdate(t *testing.T) {
	chdir(t)
	ctx := context.Background()
	h := newTestHandler("handler")

	// First run
	r1, _ := framework.NewReconciler()
	framework.Register(r1, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "item", Value: 1}}, nil
		},
	)
	if err := r1.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// Reset tracking
	h.created = make(map[string]testOutput)
	h.destroyed = make(map[string]bool)

	// Second run — different value → hash change → destroy + create
	r2, _ := framework.NewReconciler()
	framework.Register(r2, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "item", Value: 2}}, nil
		},
	)
	if err := r2.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if !h.destroyed["item"] {
		t.Error("old resource should have been destroyed on hash change")
	}
	if _, ok := h.created["item"]; !ok {
		t.Error("new resource should have been created on hash change")
	}
}

func TestNoHashChangeIsNoop(t *testing.T) {
	chdir(t)
	ctx := context.Background()
	h := newTestHandler("handler")

	// First run
	r1, _ := framework.NewReconciler()
	framework.Register(r1, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "item", Value: 1}}, nil
		},
	)
	if err := r1.Start(ctx); err != nil {
		t.Fatal(err)
	}

	h.created = make(map[string]testOutput)
	h.destroyed = make(map[string]bool)

	// Second run — same input → no-op
	r2, _ := framework.NewReconciler()
	framework.Register(r2, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "item", Value: 1}}, nil
		},
	)
	if err := r2.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if len(h.destroyed) > 0 || len(h.created) > 0 {
		t.Error("identical input should not trigger any create/destroy")
	}
}

func TestResourceRemoved(t *testing.T) {
	chdir(t)
	ctx := context.Background()
	h := newTestHandler("handler")

	// First run — two resources
	r1, _ := framework.NewReconciler()
	framework.Register(r1, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "a"}, {Name: "b"}}, nil
		},
	)
	if err := r1.Start(ctx); err != nil {
		t.Fatal(err)
	}

	h.created = make(map[string]testOutput)
	h.destroyed = make(map[string]bool)

	// Second run — only one resource → "b" should be removed
	r2, _ := framework.NewReconciler()
	framework.Register(r2, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "a"}}, nil
		},
	)
	if err := r2.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if !h.destroyed["b"] {
		t.Error("resource 'b' should have been destroyed when removed from desired list")
	}
	if h.destroyed["a"] {
		t.Error("resource 'a' should NOT have been destroyed")
	}
}

func TestResourceAdded(t *testing.T) {
	chdir(t)
	ctx := context.Background()
	h := newTestHandler("handler")

	// First run — one resource
	r1, _ := framework.NewReconciler()
	framework.Register(r1, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "a"}}, nil
		},
	)
	if err := r1.Start(ctx); err != nil {
		t.Fatal(err)
	}

	h.created = make(map[string]testOutput)

	// Second run — add another
	r2, _ := framework.NewReconciler()
	framework.Register(r2, h,
		func(_ *framework.OutputRegistry) ([]testInput, error) {
			return []testInput{{Name: "a"}, {Name: "b"}}, nil
		},
	)
	if err := r2.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if _, ok := h.created["b"]; !ok {
		t.Error("resource 'b' should have been created")
	}
	if _, ok := h.created["a"]; ok {
		t.Error("resource 'a' should NOT have been recreated (unchanged)")
	}
}
