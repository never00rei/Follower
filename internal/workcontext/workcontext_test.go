package workcontext

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCreatePersistsContextCurrentAndHistory(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	timestamp := time.Date(2026, 4, 25, 10, 30, 0, 0, time.UTC)
	ctx, err := Create("DEV-7", "ctx_parent", ContextSourceDistract, timestamp)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if !strings.HasPrefix(ctx.ID, "ctx_") {
		t.Fatalf("ctx.ID = %q, want ctx_ prefix", ctx.ID)
	}

	loaded, err := Load(ctx.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.JiraTicketID != "DEV-7" {
		t.Fatalf("loaded.JiraTicketID = %q, want %q", loaded.JiraTicketID, "DEV-7")
	}

	if len(loaded.Events) != 1 {
		t.Fatalf("len(loaded.Events) = %d, want %d", len(loaded.Events), 1)
	}

	event := loaded.Events[0]
	if event.Type != EventContextCreated {
		t.Fatalf("event.Type = %q, want %q", event.Type, EventContextCreated)
	}

	if !strings.HasPrefix(event.ID, "evt_") {
		t.Fatalf("event.ID = %q, want evt_ prefix", event.ID)
	}

	data, ok := event.Data.(ContextCreatedData)
	if !ok {
		t.Fatalf("event.Data type = %T, want ContextCreatedData", event.Data)
	}

	if data.Source != ContextSourceDistract {
		t.Fatalf("data.Source = %q, want %q", data.Source, ContextSourceDistract)
	}

	current, err := LoadCurrent()
	if err != nil {
		t.Fatalf("LoadCurrent() error = %v", err)
	}

	if current.ContextID != ctx.ID {
		t.Fatalf("current.ContextID = %q, want %q", current.ContextID, ctx.ID)
	}

	history, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}

	if len(history.Contexts) != 1 {
		t.Fatalf("len(history.Contexts) = %d, want %d", len(history.Contexts), 1)
	}

	if history.Contexts[0].ContextID != ctx.ID {
		t.Fatalf("history.Contexts[0].ContextID = %q, want %q", history.Contexts[0].ContextID, ctx.ID)
	}
}

func TestLoadHistoryReturnsEmptyHistoryWhenMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	history, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}

	if len(history.Contexts) != 0 {
		t.Fatalf("len(history.Contexts) = %d, want %d", len(history.Contexts), 0)
	}
}

func TestLoadCurrentReturnsNotExistWhenMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := LoadCurrent()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadCurrent() error = %v, want os.ErrNotExist", err)
	}
}

func TestRecordCheckpointAppendsTypedEvent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	createdAt := time.Date(2026, 4, 25, 10, 30, 0, 0, time.UTC)
	checkpointAt := createdAt.Add(15 * time.Minute)

	ctx, err := Create("DEV-7", "", ContextSourceFollow, createdAt)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := RecordCheckpoint(ctx.ID, "Implemented context storage", CheckpointTargets{Jira: true}, checkpointAt)
	if err != nil {
		t.Fatalf("RecordCheckpoint() error = %v", err)
	}

	if len(updated.Events) != 2 {
		t.Fatalf("len(updated.Events) = %d, want %d", len(updated.Events), 2)
	}

	event := updated.Events[1]
	if event.Type != EventCheckpointRecorded {
		t.Fatalf("event.Type = %q, want %q", event.Type, EventCheckpointRecorded)
	}

	data, ok := event.Data.(CheckpointRecordedData)
	if !ok {
		t.Fatalf("event.Data type = %T, want CheckpointRecordedData", event.Data)
	}

	if data.Message != "Implemented context storage" {
		t.Fatalf("data.Message = %q, want %q", data.Message, "Implemented context storage")
	}

	if !data.Targets.Jira {
		t.Fatal("data.Targets.Jira = false, want true")
	}

	loaded, err := Load(ctx.ID)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if _, ok := loaded.Events[1].Data.(CheckpointRecordedData); !ok {
		t.Fatalf("loaded.Events[1].Data type = %T, want CheckpointRecordedData", loaded.Events[1].Data)
	}

	current, err := LoadCurrent()
	if err != nil {
		t.Fatalf("LoadCurrent() error = %v", err)
	}

	if !current.UpdatedAt.Equal(checkpointAt) {
		t.Fatalf("current.UpdatedAt = %s, want %s", current.UpdatedAt, checkpointAt)
	}
}
