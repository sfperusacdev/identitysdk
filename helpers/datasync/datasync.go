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
	Insert func(ctx context.Context, y Y) error
	Update func(ctx context.Context, y Y) error
	Delete func(ctx context.Context, y Y) error
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
		mapped := strategy.Map(ext)
		matched := -1

		for i, loc := range local {
			if strategy.Equals(ext, loc) {
				matched = i
				break
			}
		}

		if matched >= 0 {
			if strategy.Update != nil {
				if err := strategy.Update(ctx, mapped); err != nil {
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
			if err := strategy.Insert(ctx, mapped); err != nil {
				return result, err
			}
			result.Inserted++
			continue
		}
	}

	if strategy.Delete != nil {
		for i, loc := range local {
			if used[i] {
				continue
			}
			if err := strategy.Delete(ctx, loc); err != nil {
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
	InsertBatch func(ctx context.Context, ys []Y) error
	UpdateBatch func(ctx context.Context, ys []Y) error
	DeleteBatch func(ctx context.Context, ys []Y) error
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
	updateItems := make([]Y, 0)
	deleteItems := make([]Y, 0)

	for _, ext := range external {
		mapped := strategy.Map(ext)
		matched := -1

		for i, loc := range local {
			if strategy.Equals(ext, loc) {
				matched = i
				break
			}
		}

		if matched >= 0 {
			updateItems = append(updateItems, mapped)
			used[matched] = true
			continue
		}

		newItems = append(newItems, mapped)
	}

	for i, loc := range local {
		if !used[i] {
			deleteItems = append(deleteItems, loc)
		}
	}

	if len(newItems) > 0 && strategy.InsertBatch != nil {
		if err := strategy.InsertBatch(ctx, newItems); err != nil {
			return result, err
		}
		result.Inserted = len(newItems)
	}

	if len(updateItems) > 0 && strategy.UpdateBatch != nil {
		if err := strategy.UpdateBatch(ctx, updateItems); err != nil {
			return result, err
		}
		result.Updated = len(updateItems)
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
