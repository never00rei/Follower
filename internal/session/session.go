package session

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/never00rei/Follower/internal/config"
)

const (
	activeSessionFile = "active_session.json"
)

func Follow(issueID string, w io.Writer) error {
	issueID = strings.TrimSpace(issueID)
	if issueID == "" {
		return errors.New("issue id is required")
	}

	current, err := loadActiveSession()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if current != nil {
		return fmt.Errorf("an active session already exists for %s", current.IssueID)
	}

	if w != nil {
		if _, err := fmt.Fprintf(w, "Starting session for %s\n", issueID); err != nil {
			return err
		}
	}

	now := time.Now().UTC()
	s := &Session{
		ID:        now.Format("20060102T150405.000000000Z07:00"),
		IssueID:   issueID,
		StartedAt: now,
	}

	if err := saveActiveSession(s); err != nil {
		return err
	}

	if err := runShell(s); err != nil {
		return err
	}

	return Done()
}

func AddCheckpoint(message string) error {
	s, err := requireActiveSession()
	if err != nil {
		return err
	}

	if strings.TrimSpace(message) == "" {
		message, err = openEditor(s)
		if err != nil {
			return err
		}
	}

	message = trimEditorContent(message)
	if message == "" {
		return errors.New("checkpoint message is empty")
	}

	s.Checkpoints = append(s.Checkpoints, Checkpoint{
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})

	return saveActiveSession(s)
}

func Status(w io.Writer) error {
	s, err := loadActiveSession()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, writeErr := fmt.Fprintln(w, "No active session.")
			return writeErr
		}

		return err
	}

	_, err = fmt.Fprintf(
		w,
		"Active issue: %s\nSession ID: %s\nStarted: %s\nElapsed: %s\nCheckpoints: %d\n",
		s.IssueID,
		s.ID,
		s.StartedAt.Format(time.RFC3339),
		time.Since(s.StartedAt).Round(time.Second),
		len(s.Checkpoints),
	)
	return err
}

func Done() error {
	s, err := loadActiveSession()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("no active session")
		}

		return err
	}

	now := time.Now().UTC()
	s.EndedAt = &now

	if err := archiveSession(s); err != nil {
		return err
	}

	return os.Remove(activeSessionPath())
}

func runShell(s *Session) error {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/sh"
	}

	cmd := exec.Command(shellPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"FOLLOWER_ISSUE_ID="+s.IssueID,
		"FOLLOWER_SESSION_ID="+s.ID,
		"FOLLOWER_SESSION_STARTED_AT="+s.StartedAt.Format(time.RFC3339),
	)

	return cmd.Run()
}

func requireActiveSession() (*Session, error) {
	s, err := loadActiveSession()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("no active session")
		}

		return nil, err
	}

	return s, nil
}

func loadActiveSession() (*Session, error) {
	data, err := os.ReadFile(activeSessionPath())
	if err != nil {
		return nil, err
	}

	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func saveActiveSession(s *Session) error {
	if _, err := config.EnsureDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(activeSessionPath(), data, 0o600)
}

func archiveSession(s *Session) error {
	dir, err := config.Dir()
	if err != nil {
		return err
	}

	sessionsDir := filepath.Join(dir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(sessionsDir, s.ID+".json"), data, 0o600)
}

func activeSessionPath() string {
	dir, err := config.Dir()
	if err != nil {
		return activeSessionFile
	}

	return filepath.Join(dir, activeSessionFile)
}

func openEditor(s *Session) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = fallbackEditor()
	}

	tmpFile, err := os.CreateTemp("", "follower-checkpoint-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	template := checkpointTemplate(s)
	if _, err := tmpFile.WriteString(template); err != nil {
		tmpFile.Close()
		return "", err
	}

	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func fallbackEditor() string {
	for _, editor := range []string{"vim", "nano", "vi"} {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return "vi"
}

func checkpointTemplate(s *Session) string {
	return fmt.Sprintf(
		"# Follower checkpoint\n# Issue: %s\n# Session: %s\n# Lines starting with # are ignored\n\n",
		s.IssueID,
		s.ID,
	)
}

func trimEditorContent(content string) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		lines = append(lines, line)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
