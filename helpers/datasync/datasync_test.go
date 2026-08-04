package datasync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sfperusacdev/identitysdk/helpers/datasync"
)

type Ext struct {
	ID   string
	Val  int
	Name string
}

type Loc struct {
	ID   string
	Val  int
	Name string
}

func TestSyncBasic(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1", Name: "A"}, {ID: "2", Name: "B"}}
	local := []Loc{{ID: "1", Name: "Old A"}, {ID: "3", Name: "C"}}

	ins := 0
	up := 0
	del := 0

	strategy := datasync.SyncStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		Insert: func(ctx context.Context, newY Loc) error {
			ins++
			return nil
		},
		Update: func(ctx context.Context, oldY Loc, newY Loc) error {
			up++
			return nil
		},
		Delete: func(ctx context.Context, oldY Loc) error {
			del++
			return nil
		},
	}

	r, err := datasync.Sync(ctx, external, local, strategy)
	if err != nil {
		t.Fatal(err)
	}

	if r.Inserted != ins || r.Updated != up || r.Deleted != del {
		t.Fatal("unexpected result counts")
	}
}

func TestSyncBatchBasic(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1", Name: "A"}, {ID: "2", Name: "B"}}
	local := []Loc{{ID: "1", Name: "Old A"}, {ID: "3", Name: "C"}}

	ins := 0
	up := 0
	del := 0

	strategy := datasync.SyncBatchStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		InsertBatch: func(ctx context.Context, newYs []Loc) error {
			ins = len(newYs)
			return nil
		},
		UpdateBatch: func(ctx context.Context, oldYs []Loc, newYs []Loc) error {
			up = len(oldYs)
			return nil
		},
		DeleteBatch: func(ctx context.Context, oldYs []Loc) error {
			del = len(oldYs)
			return nil
		},
	}

	r, err := datasync.SyncBatch(ctx, external, local, strategy)
	if err != nil {
		t.Fatal(err)
	}

	if r.Inserted != ins || r.Updated != up || r.Deleted != del {
		t.Fatal("unexpected result counts")
	}
}

func TestSyncMissingRequired(t *testing.T) {
	ctx := context.Background()
	external := []Ext{{ID: "1"}}
	local := []Loc{{ID: "1"}}

	_, err := datasync.Sync(ctx, external, local, datasync.SyncStrategy[Ext, Loc]{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSyncBatchMissingRequired(t *testing.T) {
	ctx := context.Background()
	external := []Ext{{ID: "1"}}
	local := []Loc{{ID: "1"}}

	_, err := datasync.SyncBatch(ctx, external, local, datasync.SyncBatchStrategy[Ext, Loc]{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSyncErrorPropagation(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1", Name: "A"}}
	local := []Loc{{ID: "1", Name: "A"}}

	strategy := datasync.SyncStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		Update: func(ctx context.Context, oldY Loc, newY Loc) error {
			return errors.New("update error")
		},
	}

	_, err := datasync.Sync(ctx, external, local, strategy)
	if err == nil || err.Error() != "update error" {
		t.Fatal("unexpected error")
	}
}

func TestSyncBatchErrorPropagation(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1"}, {ID: "2"}}
	local := []Loc{{ID: "1"}, {ID: "3"}}

	strategy := datasync.SyncBatchStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID} },
		InsertBatch: func(ctx context.Context, newYs []Loc) error {
			return errors.New("insert batch error")
		},
	}

	_, err := datasync.SyncBatch(ctx, external, local, strategy)
	if err == nil || err.Error() != "insert batch error" {
		t.Fatal("unexpected error")
	}
}

func TestSyncNoChanges(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1", Val: 10, Name: "A"}, {ID: "2", Val: 20, Name: "B"}}
	local := []Loc{{ID: "1", Val: 10, Name: "A"}, {ID: "2", Val: 20, Name: "B"}}

	strategy := datasync.SyncStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Val: e.Val, Name: e.Name} },
	}

	r, err := datasync.Sync(ctx, external, local, strategy)
	if err != nil {
		t.Fatal(err)
	}

	if r.Inserted != 0 || r.Updated != 0 || r.Deleted != 0 || r.Unchanged != 2 {
		t.Fatal("unexpected counts")
	}
}

func TestSyncBatchNoChanges(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1", Val: 10, Name: "A"}, {ID: "2", Val: 20, Name: "B"}}
	local := []Loc{{ID: "1", Val: 10, Name: "A"}, {ID: "2", Val: 20, Name: "B"}}

	strategy := datasync.SyncBatchStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Val: e.Val, Name: e.Name} },
	}

	r, err := datasync.SyncBatch(ctx, external, local, strategy)
	if err != nil {
		t.Fatal(err)
	}

	if r.Inserted != 0 || r.Updated != 0 || r.Deleted != 0 || r.Unchanged != 2 {
		t.Fatal("unexpected counts")
	}
}

func TestSyncChangedSkipsUnchangedMatches(t *testing.T) {
	local := []Loc{{ID: "1", Name: "same"}, {ID: "2", Name: "old"}}
	updates := 0
	deletes := 0

	result, err := datasync.Sync(context.Background(), []Ext{
		{ID: "1", Name: "same"},
		{ID: "2", Name: "new"},
	}, local, datasync.SyncStrategy[Ext, Loc]{
		Equals:  func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:     func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		Changed: func(old Loc, new Loc) bool { return old.Name != new.Name },
		Update: func(context.Context, Loc, Loc) error {
			updates++
			return nil
		},
		Delete: func(context.Context, Loc) error {
			deletes++
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Updated != 1 || updates != 1 || result.Unchanged != 1 {
		t.Fatalf("unexpected update counts: result=%+v updates=%d", result, updates)
	}
	if result.Deleted != 0 || deletes != 0 {
		t.Fatalf("unchanged match was treated as deleted: result=%+v deletes=%d", result, deletes)
	}
}

func TestSyncBatchChangedFiltersUpdates(t *testing.T) {
	var oldValues []Loc
	var newValues []Loc

	result, err := datasync.SyncBatch(context.Background(), []Ext{
		{ID: "1", Name: "same"},
		{ID: "2", Name: "new"},
	}, []Loc{{ID: "1", Name: "same"}, {ID: "2", Name: "old"}}, datasync.SyncBatchStrategy[Ext, Loc]{
		Equals:  func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:     func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		Changed: func(old Loc, new Loc) bool { return old.Name != new.Name },
		UpdateBatch: func(_ context.Context, old []Loc, new []Loc) error {
			oldValues = old
			newValues = new
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Updated != 1 || len(oldValues) != 1 || len(newValues) != 1 {
		t.Fatalf("unexpected batch update counts: result=%+v old=%v new=%v", result, oldValues, newValues)
	}
	if oldValues[0].ID != "2" || newValues[0].Name != "new" || result.Unchanged != 1 {
		t.Fatalf("unexpected batch values: result=%+v old=%v new=%v", result, oldValues, newValues)
	}
}

func TestSync_UpdateReceivesOldAndNew(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1", Val: 100, Name: "NEW"}}
	local := []Loc{{ID: "1", Val: 10, Name: "OLD"}}

	receivedOld := Loc{}
	receivedNew := Loc{}

	strategy := datasync.SyncStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Val: e.Val, Name: e.Name} },
		Update: func(ctx context.Context, oldY Loc, newY Loc) error {
			receivedOld = oldY
			receivedNew = newY
			return nil
		},
	}

	_, err := datasync.Sync(ctx, external, local, strategy)
	if err != nil {
		t.Fatal(err)
	}

	if receivedOld.ID != "1" || receivedOld.Val != 10 || receivedOld.Name != "OLD" {
		t.Fatalf("unexpected old value: %+v", receivedOld)
	}

	if receivedNew.ID != "1" || receivedNew.Val != 100 || receivedNew.Name != "NEW" {
		t.Fatalf("unexpected new value: %+v", receivedNew)
	}
}

func TestSyncBatch_UpdateReceivesOldAndNew(t *testing.T) {
	ctx := context.Background()

	external := []Ext{{ID: "1", Val: 100, Name: "NEW"}}
	local := []Loc{{ID: "1", Val: 10, Name: "OLD"}}

	var oldValues []Loc
	var newValues []Loc

	strategy := datasync.SyncBatchStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Val: e.Val, Name: e.Name} },
		UpdateBatch: func(ctx context.Context, oldYs []Loc, newYs []Loc) error {
			oldValues = oldYs
			newValues = newYs
			return nil
		},
	}

	_, err := datasync.SyncBatch(ctx, external, local, strategy)
	if err != nil {
		t.Fatal(err)
	}

	if len(oldValues) != 1 {
		t.Fatalf("expected 1 old value, got %d", len(oldValues))
	}

	if len(newValues) != 1 {
		t.Fatalf("expected 1 new value, got %d", len(newValues))
	}

	oldY := oldValues[0]
	newY := newValues[0]

	if oldY.ID != "1" || oldY.Val != 10 || oldY.Name != "OLD" {
		t.Fatalf("unexpected old value: %+v", oldY)
	}

	if newY.ID != "1" || newY.Val != 100 || newY.Name != "NEW" {
		t.Fatalf("unexpected new value: %+v", newY)
	}
}

func TestSync_DoesNotReuseLocalMatch(t *testing.T) {
	external := []Ext{{ID: "1", Name: "first"}, {ID: "1", Name: "second"}}
	local := []Loc{{ID: "1", Name: "old"}}
	inserted := 0
	updated := 0

	result, err := datasync.Sync(context.Background(), external, local, datasync.SyncStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		Insert: func(context.Context, Loc) error {
			inserted++
			return nil
		},
		Update: func(context.Context, Loc, Loc) error {
			updated++
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Inserted != 1 || inserted != 1 {
		t.Fatalf("expected one insert, result=%+v callback=%d", result, inserted)
	}
	if result.Updated != 1 || updated != 1 {
		t.Fatalf("expected one update, result=%+v callback=%d", result, updated)
	}
}

func TestSyncBatch_DoesNotReuseLocalMatch(t *testing.T) {
	external := []Ext{{ID: "1", Name: "first"}, {ID: "1", Name: "second"}}
	local := []Loc{{ID: "1", Name: "old"}}
	var inserted []Loc
	var updatedNew []Loc

	result, err := datasync.SyncBatch(context.Background(), external, local, datasync.SyncBatchStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		InsertBatch: func(_ context.Context, values []Loc) error {
			inserted = values
			return nil
		},
		UpdateBatch: func(_ context.Context, _, values []Loc) error {
			updatedNew = values
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Inserted != 1 || len(inserted) != 1 {
		t.Fatalf("expected one insert, result=%+v values=%v", result, inserted)
	}
	if result.Updated != 1 || len(updatedNew) != 1 {
		t.Fatalf("expected one update, result=%+v values=%v", result, updatedNew)
	}
}

func TestSyncBatch_CountsUnchangedWithPartialStrategy(t *testing.T) {
	external := []Ext{{ID: "1", Name: "same"}, {ID: "2", Name: "new"}}
	local := []Loc{{ID: "1", Name: "same"}}
	inserted := 0

	result, err := datasync.SyncBatch(context.Background(), external, local, datasync.SyncBatchStrategy[Ext, Loc]{
		Equals: func(e Ext, l Loc) bool { return e.ID == l.ID },
		Map:    func(e Ext) Loc { return Loc{ID: e.ID, Name: e.Name} },
		InsertBatch: func(_ context.Context, values []Loc) error {
			inserted = len(values)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Inserted != 1 || inserted != 1 {
		t.Fatalf("expected one insert, result=%+v callback=%d", result, inserted)
	}
	if result.Unchanged != 1 {
		t.Fatalf("expected one unchanged item, result=%+v", result)
	}
}
