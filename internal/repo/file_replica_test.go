package repo

import (
	"context"
	"testing"
	"time"

	"github.com/amigoer/kite/internal/model"
)

func TestFileReplicaRepo_CreateAndList(t *testing.T) {
	r := NewFileReplicaRepo(newTestDB(t))
	ctx := context.Background()

	rep := &model.FileReplica{
		ID:              "rep1",
		FileID:          "f1",
		StorageConfigID: "cfg1",
		Status:          model.ReplicaStatusPending,
	}
	if err := r.Create(ctx, rep); err != nil {
		t.Fatalf("Create: %v", err)
	}

	list, err := r.ListByFile(ctx, "f1")
	if err != nil || len(list) != 1 || list[0].ID != "rep1" {
		t.Fatalf("ListByFile: err=%v len=%d", err, len(list))
	}
}

func TestFileReplicaRepo_ListByFile_Empty(t *testing.T) {
	r := NewFileReplicaRepo(newTestDB(t))
	list, err := r.ListByFile(context.Background(), "nobody")
	if err != nil || len(list) != 0 {
		t.Fatalf("ListByFile empty: err=%v len=%d", err, len(list))
	}
}

func TestFileReplicaRepo_UpdateStatus(t *testing.T) {
	r := NewFileReplicaRepo(newTestDB(t))
	ctx := context.Background()

	r.Create(ctx, &model.FileReplica{
		ID: "rep2", FileID: "f1", StorageConfigID: "cfg1",
		Status: model.ReplicaStatusPending,
	})

	if err := r.UpdateStatus(ctx, "rep2", model.ReplicaStatusFailed, "connection timeout"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	list, _ := r.ListByFile(ctx, "f1")
	if list[0].Status != model.ReplicaStatusFailed {
		t.Fatalf("status not updated: %s", list[0].Status)
	}
	if list[0].ErrorMsg != "connection timeout" {
		t.Fatalf("error_msg not updated: %s", list[0].ErrorMsg)
	}
}

func TestFileReplicaRepo_DeleteByFile(t *testing.T) {
	r := NewFileReplicaRepo(newTestDB(t))
	ctx := context.Background()

	r.Create(ctx, &model.FileReplica{ID: "rep3", FileID: "fA", StorageConfigID: "cfg1", Status: model.ReplicaStatusOK})
	r.Create(ctx, &model.FileReplica{ID: "rep4", FileID: "fA", StorageConfigID: "cfg2", Status: model.ReplicaStatusOK})
	r.Create(ctx, &model.FileReplica{ID: "rep5", FileID: "fB", StorageConfigID: "cfg1", Status: model.ReplicaStatusOK})

	if err := r.DeleteByFile(ctx, "fA"); err != nil {
		t.Fatalf("DeleteByFile: %v", err)
	}

	listA, _ := r.ListByFile(ctx, "fA")
	if len(listA) != 0 {
		t.Fatalf("fA replicas should be deleted, got %d", len(listA))
	}

	listB, _ := r.ListByFile(ctx, "fB")
	if len(listB) != 1 {
		t.Fatalf("fB replicas should be untouched, got %d", len(listB))
	}
}

// TestFileReplicaRepo_MarkStalePending_FlipsOnlyOldPending covers the
// startup reconciliation path: only pending rows past the threshold flip
// to failed; OK/failed rows and fresh pending rows stay untouched.
func TestFileReplicaRepo_MarkStalePending_FlipsOnlyOldPending(t *testing.T) {
	db := newTestDB(t)
	r := NewFileReplicaRepo(db)
	ctx := context.Background()

	// Fresh pending (younger than threshold) — must NOT flip.
	if err := r.Create(ctx, &model.FileReplica{
		ID: "fresh-pending", FileID: "f1", StorageConfigID: "cfg1",
		Status: model.ReplicaStatusPending,
	}); err != nil {
		t.Fatalf("seed fresh: %v", err)
	}

	// Old pending (older than threshold) — must flip.
	if err := r.Create(ctx, &model.FileReplica{
		ID: "stale-pending", FileID: "f1", StorageConfigID: "cfg2",
		Status: model.ReplicaStatusPending,
	}); err != nil {
		t.Fatalf("seed stale: %v", err)
	}
	// Forcibly age the row so the WHERE clause hits it. GORM's autoUpdateTime
	// keeps writing "now" for us otherwise.
	old := time.Now().Add(-2 * time.Hour)
	if err := db.Model(&model.FileReplica{}).
		Where("id = ?", "stale-pending").
		UpdateColumn("updated_at", old).Error; err != nil {
		t.Fatalf("age stale row: %v", err)
	}

	// Old ok — must NOT flip (status already terminal).
	if err := r.Create(ctx, &model.FileReplica{
		ID: "stale-ok", FileID: "f1", StorageConfigID: "cfg3",
		Status: model.ReplicaStatusOK,
	}); err != nil {
		t.Fatalf("seed ok: %v", err)
	}
	if err := db.Model(&model.FileReplica{}).
		Where("id = ?", "stale-ok").
		UpdateColumn("updated_at", old).Error; err != nil {
		t.Fatalf("age ok row: %v", err)
	}

	flipped, err := r.MarkStalePending(ctx, time.Hour)
	if err != nil {
		t.Fatalf("MarkStalePending: %v", err)
	}
	if flipped != 1 {
		t.Fatalf("expected exactly 1 row flipped, got %d", flipped)
	}

	// Verify the right row flipped and the others are untouched.
	all, err := r.ListByFile(ctx, "f1")
	if err != nil {
		t.Fatalf("ListByFile: %v", err)
	}
	byID := map[string]model.FileReplica{}
	for _, rep := range all {
		byID[rep.ID] = rep
	}
	if got := byID["fresh-pending"].Status; got != model.ReplicaStatusPending {
		t.Fatalf("fresh-pending status = %s, want pending", got)
	}
	if got := byID["stale-pending"].Status; got != model.ReplicaStatusFailed {
		t.Fatalf("stale-pending status = %s, want failed", got)
	}
	if got := byID["stale-pending"].ErrorMsg; got == "" {
		t.Fatalf("stale-pending error_msg should be set, got empty")
	}
	if got := byID["stale-ok"].Status; got != model.ReplicaStatusOK {
		t.Fatalf("stale-ok status = %s, want ok (should not flip)", got)
	}
}

func TestFileReplicaRepo_MarkStalePending_RejectsNonPositive(t *testing.T) {
	r := NewFileReplicaRepo(newTestDB(t))
	if _, err := r.MarkStalePending(context.Background(), 0); err == nil {
		t.Fatal("expected error for zero threshold")
	}
	if _, err := r.MarkStalePending(context.Background(), -time.Minute); err == nil {
		t.Fatal("expected error for negative threshold")
	}
}

func TestFileReplica_StatusConstants(t *testing.T) {
	if model.ReplicaStatusPending != "pending" {
		t.Fatal("wrong pending constant")
	}
	if model.ReplicaStatusOK != "ok" {
		t.Fatal("wrong ok constant")
	}
	if model.ReplicaStatusFailed != "failed" {
		t.Fatal("wrong failed constant")
	}
}
