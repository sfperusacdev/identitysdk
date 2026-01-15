package datasync

import (
	"context"
	"errors"
)

type SyncResult struct {
	Inserted  int
	Updated   int
	Deleted   int
	Unchanged int
}

type SyncStrategy[X any, Y any] struct {
	Equals func(ext X, loc Y) bool
	Map    func(ext X) Y
	Insert func(ctx context.Context, new Y) error
	Update func(ctx context.Context, old Y, new Y) error
	Delete func(ctx context.Context, old Y) error
}

func Sync[X any, Y any](
	ctx context.Context,
	external []X,
	local []Y,
	strategy SyncStrategy[X, Y],
) (SyncResult, error) {
	if strategy.Equals == nil {
		return SyncResult{}, errors.New("datasync: Equals function is required")
	}
	if strategy.Map == nil {
		return SyncResult{}, errors.New("datasync: Map function is required")
	}

	result := SyncResult{}
	used := make([]bool, len(local))

	for _, ext := range external {
		newY := strategy.Map(ext)
		matched := -1
		var oldY Y

		for i, loc := range local {
			if strategy.Equals(ext, loc) {
				matched = i
				oldY = loc
				break
			}
		}

		if matched >= 0 {
			if strategy.Update != nil {
				if err := strategy.Update(ctx, oldY, newY); err != nil {
					return result, err
				}
				result.Updated++
			} else {
				result.Unchanged++
			}
			used[matched] = true
			continue
		}

		if strategy.Insert != nil {
			if err := strategy.Insert(ctx, newY); err != nil {
				return result, err
			}
			result.Inserted++
		}
	}

	if strategy.Delete != nil {
		for i, oldY := range local {
			if used[i] {
				continue
			}
			if err := strategy.Delete(ctx, oldY); err != nil {
				return result, err
			}
			result.Deleted++
		}
	}

	return result, nil
}

type SyncBatchStrategy[X any, Y any] struct {
	Equals      func(ext X, loc Y) bool
	Map         func(ext X) Y
	InsertBatch func(ctx context.Context, newYs []Y) error
	UpdateBatch func(ctx context.Context, oldYs []Y, newYs []Y) error
	DeleteBatch func(ctx context.Context, oldYs []Y) error
}

func SyncBatch[X any, Y any](
	ctx context.Context,
	external []X,
	local []Y,
	strategy SyncBatchStrategy[X, Y],
) (SyncResult, error) {
	if strategy.Equals == nil {
		return SyncResult{}, errors.New("datasync: Equals function is required")
	}
	if strategy.Map == nil {
		return SyncResult{}, errors.New("datasync: Map function is required")
	}

	result := SyncResult{}
	used := make([]bool, len(local))
	newItems := make([]Y, 0)
	updateOld := make([]Y, 0)
	updateNew := make([]Y, 0)
	deleteItems := make([]Y, 0)

	for _, ext := range external {
		newY := strategy.Map(ext)
		matched := -1
		var oldY Y

		for i, loc := range local {
			if strategy.Equals(ext, loc) {
				matched = i
				oldY = loc
				break
			}
		}

		if matched >= 0 {
			updateOld = append(updateOld, oldY)
			updateNew = append(updateNew, newY)
			used[matched] = true
			continue
		}

		newItems = append(newItems, newY)
	}

	for i, oldY := range local {
		if !used[i] {
			deleteItems = append(deleteItems, oldY)
		}
	}

	if len(newItems) > 0 && strategy.InsertBatch != nil {
		if err := strategy.InsertBatch(ctx, newItems); err != nil {
			return result, err
		}
		result.Inserted = len(newItems)
	}

	if len(updateNew) > 0 && strategy.UpdateBatch != nil {
		if err := strategy.UpdateBatch(ctx, updateOld, updateNew); err != nil {
			return result, err
		}
		result.Updated = len(updateNew)
	}

	if len(deleteItems) > 0 && strategy.DeleteBatch != nil {
		if err := strategy.DeleteBatch(ctx, deleteItems); err != nil {
			return result, err
		}
		result.Deleted = len(deleteItems)
	}

	if strategy.InsertBatch == nil && strategy.UpdateBatch == nil && strategy.DeleteBatch == nil {
		result.Unchanged = len(external)
	}

	return result, nil
}
