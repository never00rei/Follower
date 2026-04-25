package workcontext

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestCreatePersistsContextCurrentAndHistory(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	timestamp := time.Date(2026, 4, 25, 10, 30, 0, 0, time.UTC)
	ctx, err := Create("ctx_1", "evt_1", "DEV-7", "ctx_parent", ContextSourceDistract, timestamp)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if ctx.ID != "ctx_1" {
		t.Fatalf("ctx.ID = %q, want %q", ctx.ID, "ctx_1")
	}

	loaded, err := Load("ctx_1")
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

	if current.ContextID != "ctx_1" {
		t.Fatalf("current.ContextID = %q, want %q", current.ContextID, "ctx_1")
	}

	history, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}

	if len(history.Contexts) != 1 {
		t.Fatalf("len(history.Contexts) = %d, want %d", len(history.Contexts), 1)
	}

	if history.Contexts[0].ContextID != "ctx_1" {
		t.Fatalf("history.Contexts[0].ContextID = %q, want %q", history.Contexts[0].ContextID, "ctx_1")
	}
}

func TestLoadHistoryReturnsEmptyHistoryWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	history, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}

	if len(history.Contexts) != 0 {
		t.Fatalf("len(history.Contexts) = %d, want %d", len(history.Contexts), 0)
	}
}

func TestLoadCurrentReturnsNotExistWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := LoadCurrent()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LoadCurrent() error = %v, want os.ErrNotExist", err)
	}
}
