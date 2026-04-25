package workcontext

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/never00rei/Follower/internal/config"
)

const (
	contextsDirName       = "contexts"
	currentContextFile    = "current_context.json"
	contextHistoryFile    = "context_history.json"
	contextFilePermission = 0o600
	contextDirPermission  = 0o700
)

func Save(ctx *WorkContext) error {
	if ctx == nil {
		return errors.New("work context is nil")
	}

	if ctx.ID == "" {
		return errors.New("work context id is required")
	}

	dir, err := ensureContextsDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, ctx.ID+".json"), data, contextFilePermission)
}

func Load(id string) (*WorkContext, error) {
	if id == "" {
		return nil, errors.New("work context id is required")
	}

	dir, err := contextsDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return nil, err
	}

	var ctx WorkContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

func SaveCurrent(ref CurrentContextRef) error {
	if ref.ContextID == "" {
		return errors.New("current context id is required")
	}

	dir, err := config.EnsureDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(ref, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, currentContextFile), data, contextFilePermission)
}

func LoadCurrent() (*CurrentContextRef, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(dir, currentContextFile))
	if err != nil {
		return nil, err
	}

	var ref CurrentContextRef
	if err := json.Unmarshal(data, &ref); err != nil {
		return nil, err
	}

	return &ref, nil
}

func SaveHistory(history ContextHistory) error {
	dir, err := config.EnsureDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, contextHistoryFile), data, contextFilePermission)
}

func LoadHistory() (*ContextHistory, error) {
	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(dir, contextHistoryFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ContextHistory{}, nil
		}

		return nil, err
	}

	var history ContextHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	return &history, nil
}

func UpsertHistoryEntry(entry ContextHistoryEntry) error {
	if entry.ContextID == "" {
		return errors.New("context history entry id is required")
	}

	history, err := LoadHistory()
	if err != nil {
		return err
	}

	for i := range history.Contexts {
		if history.Contexts[i].ContextID == entry.ContextID {
			history.Contexts[i] = entry
			return SaveHistory(*history)
		}
	}

	history.Contexts = append(history.Contexts, entry)
	return SaveHistory(*history)
}

func contextsDir() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, contextsDirName), nil
}

func ensureContextsDir() (string, error) {
	dir, err := contextsDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(dir, contextDirPermission); err != nil {
		return "", fmt.Errorf("could not create contexts directory: %w", err)
	}

	return dir, nil
}
