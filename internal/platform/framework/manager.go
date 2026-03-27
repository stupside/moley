package framework

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

// typedNode implements node for a specific Lifecycle type.
type typedNode[TInput any, TOutput any] struct {
	handler  Lifecycle[TInput, TOutput]
	resolver InputResolver[TInput]
	deps     []string
	inputs   []TInput // resolved at reconcile time
}

func (n *typedNode[TInput, TOutput]) name() string {
	return n.handler.Name()
}

func (n *typedNode[TInput, TOutput]) dependencies() []string {
	return n.deps
}

func (n *typedNode[TInput, TOutput]) resolve(reg *OutputRegistry) error {
	inputs, err := n.resolver(reg)
	if err != nil {
		return fmt.Errorf("failed to resolve inputs for %s: %w", n.handler.Name(), err)
	}
	n.inputs = inputs
	return nil
}

func (n *typedNode[TInput, TOutput]) newManager(lf *LockFile) *nodeManager[TInput, TOutput] {
	return &nodeManager[TInput, TOutput]{
		handler:  n.handler,
		lockFile: lf,
	}
}

func (n *typedNode[TInput, TOutput]) reconcile(ctx context.Context, lf *LockFile) error {
	return n.newManager(lf).Reconcile(ctx, n.inputs)
}

func (n *typedNode[TInput, TOutput]) stop(ctx context.Context, lf *LockFile) error {
	return n.newManager(lf).Stop(ctx, n.inputs)
}

// nodeManager manages resources of a specific type with full type safety.
type nodeManager[TInput any, TOutput any] struct {
	handler  Lifecycle[TInput, TOutput]
	lockFile *LockFile
}

// unmarshalData converts interface{} data back to typed struct using JSON marshaling.
func unmarshalData[T any](data any, target *T) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}
	if err := json.Unmarshal(jsonBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON to typed struct: %w", err)
	}
	return nil
}

// Reconcile ensures the desired resources match the actual state.
func (rm *nodeManager[TInput, TOutput]) Reconcile(ctx context.Context, desiredInputs []TInput) error {
	currentRecords := rm.verifyRecords(ctx)

	toRemove, toAdd, toUpdate := rm.computeActions(desiredInputs, currentRecords)

	var errs []error

	if err := rm.removeResources(ctx, toRemove); err != nil {
		errs = append(errs, err)
	}

	if err := rm.addResources(ctx, toAdd); err != nil {
		errs = append(errs, err)
	}

	if err := rm.updateResources(ctx, toUpdate); err != nil {
		errs = append(errs, err)
	}

	if err := rm.lockFile.Save(); err != nil {
		errs = append(errs, fmt.Errorf("failed to save lock file: %w", err))
	}

	return errors.Join(errs...)
}

type verifiedRecord[TInput any, TOutput any] struct {
	snapshot  Snapshot[TInput, TOutput]
	inputHash string
}

// verifyRecords checks each lock entry against reality, removes stale ones.
func (rm *nodeManager[TInput, TOutput]) verifyRecords(ctx context.Context) []verifiedRecord[TInput, TOutput] {
	handlerName := rm.handler.Name()
	kept := make([]LockEntry, 0, len(rm.lockFile.Entries))
	var records []verifiedRecord[TInput, TOutput]

	for _, entry := range rm.lockFile.Entries {
		if entry.HandlerName != handlerName {
			kept = append(kept, entry)
			continue
		}

		var snap Snapshot[TInput, TOutput]
		if err := unmarshalData(entry.Data, &snap); err != nil {
			logger.Debugf("Failed to unmarshal entry during verify, keeping", map[string]any{
				"handler": handlerName,
				"key":     entry.Key,
				"error":   err.Error(),
			})
			kept = append(kept, entry)
			continue
		}

		status, err := rm.handler.Check(ctx, snap.Output)

		// StatusDown and no error → stale entry, drop it
		if err == nil && status == StatusDown {
			logger.Infof("Removing stale lock entry", map[string]any{
				"handler": handlerName,
				"key":     entry.Key,
			})
			continue
		}

		// StatusUp, Unknown, or error → keep conservatively
		kept = append(kept, entry)
		records = append(records, verifiedRecord[TInput, TOutput]{
			snapshot:  snap,
			inputHash: entry.InputHash,
		})
	}

	rm.lockFile.Entries = kept
	return records
}

// computeActions determines what resources need to be added, removed, or updated.
// Change detection uses input hashing instead of Equals().
func (rm *nodeManager[TInput, TOutput]) computeActions(
	desired []TInput,
	current []verifiedRecord[TInput, TOutput],
) (
	toRemove []verifiedRecord[TInput, TOutput],
	toAdd []TInput,
	toUpdate []struct {
		newInput TInput
		old      verifiedRecord[TInput, TOutput]
	},
) {
	currentMap := make(map[string]verifiedRecord[TInput, TOutput])
	for _, record := range current {
		key := rm.handler.Key(record.snapshot.Input)
		currentMap[key] = record
	}

	desiredMap := make(map[string]TInput)
	for _, input := range desired {
		key := rm.handler.Key(input)
		desiredMap[key] = input
	}

	// Find what to remove (current but not desired)
	for key, record := range currentMap {
		if _, exists := desiredMap[key]; !exists {
			toRemove = append(toRemove, record)
		}
	}

	// Find what to add or update
	for key, desiredInput := range desiredMap {
		newHash, err := computeHash(desiredInput)
		if err != nil {
			logger.Warnf("Failed to hash input, treating as changed", map[string]any{
				"handler": rm.handler.Name(),
				"key":     key,
				"error":   err.Error(),
			})
			newHash = "" // force mismatch
		}

		if currentRecord, exists := currentMap[key]; exists {
			if newHash != currentRecord.inputHash {
				toUpdate = append(toUpdate, struct {
					newInput TInput
					old      verifiedRecord[TInput, TOutput]
				}{newInput: desiredInput, old: currentRecord})
			}
		} else {
			toAdd = append(toAdd, desiredInput)
		}
	}

	return toRemove, toAdd, toUpdate
}

// createAndVerify creates a resource and verifies it's in StatusUp.
func (rm *nodeManager[TInput, TOutput]) createAndVerify(ctx context.Context, input TInput) (TOutput, error) {
	output, err := rm.handler.Create(ctx, input)
	if err != nil {
		return output, fmt.Errorf("failed to create resource: %w", err)
	}

	if err := rm.errorIfNotUp(ctx, output); err != nil {
		return output, fmt.Errorf("failed to verify created resource: %w", err)
	}

	return output, nil
}

func (rm *nodeManager[TInput, TOutput]) errorIfNotUp(ctx context.Context, output TOutput) error {
	status, err := rm.handler.Check(ctx, output)
	if err != nil {
		return fmt.Errorf("failed to verify resource status: %w", err)
	}
	if status != StatusUp {
		return fmt.Errorf("resource not in up state (status: %s)", status)
	}
	return nil
}

func (rm *nodeManager[TInput, TOutput]) removeResources(ctx context.Context, toRemove []verifiedRecord[TInput, TOutput]) error {
	handlerName := rm.handler.Name()
	var errs []error
	for _, record := range toRemove {
		logger.Infof("Removing resource", map[string]any{
			"handler": handlerName,
		})

		if err := rm.handler.Destroy(ctx, record.snapshot.Output); err != nil {
			errs = append(errs, fmt.Errorf("failed to destroy resource: %w", err))
			continue
		}

		rm.removeFromRegistry(record.snapshot)
	}
	return errors.Join(errs...)
}

func (rm *nodeManager[TInput, TOutput]) addResources(ctx context.Context, toAdd []TInput) error {
	handlerName := rm.handler.Name()
	var errs []error
	for _, input := range toAdd {
		logger.Infof("Adding resource", map[string]any{
			"handler": handlerName,
		})

		existingOutput, status, err := rm.handler.Recover(ctx, input)

		var output TOutput

		if status == StatusUp {
			logger.Infof("Resource already exists, reusing", map[string]any{
				"handler": handlerName,
			})
			output = existingOutput
		} else {
			if status == StatusUnknown {
				logger.Warnf("Unable to check if resource exists, attempting creation", map[string]any{
					"handler": handlerName,
					"error":   err,
				})
			}

			if output, err = rm.createAndVerify(ctx, input); err != nil {
				errs = append(errs, err)
				continue
			}
		}

		rm.addToRegistry(Snapshot[TInput, TOutput]{Input: input, Output: output})
	}
	return errors.Join(errs...)
}

func (rm *nodeManager[TInput, TOutput]) updateResources(
	ctx context.Context,
	toUpdate []struct {
		newInput TInput
		old      verifiedRecord[TInput, TOutput]
	},
) error {
	handlerName := rm.handler.Name()
	var errs []error
	for _, update := range toUpdate {
		logger.Infof("Updating resource", map[string]any{
			"handler": handlerName,
		})

		if err := rm.handler.Destroy(ctx, update.old.snapshot.Output); err != nil {
			errs = append(errs, fmt.Errorf("failed to destroy old resource during update: %w", err))
			continue
		}

		newOutput, err := rm.createAndVerify(ctx, update.newInput)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create updated resource: %w", err))
			continue
		}

		rm.removeFromRegistry(update.old.snapshot)
		rm.addToRegistry(Snapshot[TInput, TOutput]{
			Input:  update.newInput,
			Output: newOutput,
		})
	}
	return errors.Join(errs...)
}

// addToRegistry and removeFromRegistry mutate in-memory only. Save() is called once at the end of Reconcile/Stop.
func (rm *nodeManager[TInput, TOutput]) addToRegistry(snap Snapshot[TInput, TOutput]) {
	inputHash, _ := computeHash(snap.Input)

	entry := LockEntry{
		Key:         rm.handler.Key(snap.Input),
		Data:        snap,
		HandlerName: rm.handler.Name(),
		InputHash:   inputHash,
	}

	for i, e := range rm.lockFile.Entries {
		if e.Key == entry.Key && e.HandlerName == entry.HandlerName {
			rm.lockFile.Entries[i] = entry
			return
		}
	}
	rm.lockFile.Entries = append(rm.lockFile.Entries, entry)
}

func (rm *nodeManager[TInput, TOutput]) removeFromRegistry(snap Snapshot[TInput, TOutput]) {
	key := rm.handler.Key(snap.Input)
	handlerName := rm.handler.Name()

	for i, e := range rm.lockFile.Entries {
		if e.Key == key && e.HandlerName == handlerName {
			rm.lockFile.Entries = append(rm.lockFile.Entries[:i], rm.lockFile.Entries[i+1:]...)
			return
		}
	}
}

// Stop removes all resources managed by this handler (tracked + recovered).
func (rm *nodeManager[TInput, TOutput]) Stop(ctx context.Context, inputs []TInput) error {
	currentRecords := rm.verifyRecords(ctx)

	trackedMap := make(map[string]verifiedRecord[TInput, TOutput])
	for _, record := range currentRecords {
		key := rm.handler.Key(record.snapshot.Input)
		trackedMap[key] = record
	}

	allToRemove := currentRecords

	handlerName := rm.handler.Name()
	for _, input := range inputs {
		key := rm.handler.Key(input)
		if _, exists := trackedMap[key]; exists {
			continue
		}

		output, status, err := rm.handler.Recover(ctx, input)
		switch status {
		case StatusUp:
			logger.Infof("Found untracked running resource", map[string]any{
				"handler": handlerName,
			})
			allToRemove = append(allToRemove, verifiedRecord[TInput, TOutput]{
				snapshot: Snapshot[TInput, TOutput]{
					Input:  input,
					Output: output,
				},
			})
		case StatusUnknown:
			logger.Warnf("Unable to determine untracked resource state, skipping", map[string]any{
				"handler": handlerName,
				"error":   err,
			})
		}
	}

	logger.Debugf("Total resources to remove", map[string]any{
		"handler":      handlerName,
		"remove_count": len(allToRemove),
	})

	err := rm.removeResources(ctx, allToRemove)

	if saveErr := rm.lockFile.Save(); saveErr != nil {
		err = errors.Join(err, fmt.Errorf("failed to save lock file: %w", saveErr))
	}

	return err
}
