package session

import (
	"bufio"
	"context"
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
	"github.com/never00rei/Follower/internal/git"
	"github.com/never00rei/Follower/internal/jira"
	"github.com/never00rei/Follower/internal/workcontext"
)

const (
	activeSessionFile = "active_session.json"
)

func Follow(input FollowInput, w io.Writer) error {
	input.IssueID = strings.TrimSpace(input.IssueID)
	input.WorkContextID = strings.TrimSpace(input.WorkContextID)
	input.ParentContextID = strings.TrimSpace(input.ParentContextID)

	if input.IssueID == "" {
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
		if _, err := fmt.Fprintf(w, "Starting session for %s\n", input.IssueID); err != nil {
			return err
		}
	}

	now := input.StartedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	s := &Session{
		ID:              now.Format("20060102T150405.000000000Z07:00"),
		IssueID:         input.IssueID,
		WorkContextID:   input.WorkContextID,
		ParentContextID: input.ParentContextID,
		StartedAt:       now,
	}

	if err := saveActiveSession(s); err != nil {
		return err
	}

	if err := runShell(s); err != nil {
		return err
	}

	return Done()
}

func EnsureNoActive() error {
	current, err := loadActiveSession()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	return fmt.Errorf("an active session already exists for %s", current.IssueID)
}

func AddCheckpoint(message string) error {
	return AddPreparedCheckpoint(message, CheckpointTargets{})
}

func AddPreparedCheckpoint(message string, targets CheckpointTargets) error {
	s, err := loadOptionalActiveSession()
	if err != nil {
		return err
	}

	createdAt := time.Now().UTC()

	ctx, err := resolveCheckpointContext(s)
	if err != nil {
		return err
	}

	message, gitSummary, err := collectCheckpointInput(ctx.JiraTicketID, checkpointTrackingID(ctx, s), message, targets, createdAt)
	if err != nil {
		return err
	}

	if message == "" {
		return errors.New("checkpoint message is empty")
	}

	checkpoint := Checkpoint{
		Message:   message,
		CreatedAt: createdAt,
		Targets:   targets,
	}

	if targets.Git {
		checkpoint.Git = &GitCheckpoint{
			CommitMessage: gitSummary,
			CommitTime:    createdAt,
			CommitBody:    message,
		}

		commitSHA, err := commitGitCheckpoint(checkpoint.Git)
		if err != nil {
			return err
		}

		checkpoint.Git.CommitSHA = commitSHA
	}

	if targets.Jira {
		if s != nil {
			checkpoint.Jira = &JiraCheckpoint{
				WindowStartedAt: checkpointWindowStart(s),
				WindowEndedAt:   createdAt,
			}
		}
	}

	if _, err := workcontext.RecordCheckpoint(
		ctx.ID,
		message,
		workcontext.CheckpointTargets{
			Jira: targets.Jira,
			Git:  targets.Git,
		},
		createdAt,
	); err != nil {
		return err
	}

	if s == nil {
		return nil
	}

	s.Checkpoints = append(s.Checkpoints, checkpoint)

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
		"Active issue: %s\nSession ID: %s\nWork context ID: %s\nStarted: %s\nElapsed: %s\nCheckpoints: %d\n",
		s.IssueID,
		s.ID,
		s.WorkContextID,
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
		"FOLLOWER_JIRA_TICKET_ID="+s.IssueID,
		"FOLLOWER_SESSION_ID="+s.ID,
		"FOLLOWER_SESSION_STARTED_AT="+s.StartedAt.Format(time.RFC3339),
	)

	if s.WorkContextID != "" {
		cmd.Env = append(cmd.Env, "FOLLOWER_CONTEXT_ID="+s.WorkContextID)
	}

	if s.ParentContextID != "" {
		cmd.Env = append(cmd.Env, "FOLLOWER_PARENT_CONTEXT_ID="+s.ParentContextID)
	}

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

func loadOptionalActiveSession() (*Session, error) {
	s, err := loadActiveSession()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	return s, nil
}

func resolveCheckpointContext(s *Session) (*workcontext.WorkContext, error) {
	contextID := strings.TrimSpace(os.Getenv("FOLLOWER_CONTEXT_ID"))
	if contextID != "" {
		return workcontext.Load(contextID)
	}

	current, err := workcontext.LoadCurrent()
	if err == nil {
		return workcontext.Load(current.ContextID)
	}

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if s != nil && s.WorkContextID != "" {
		return workcontext.Load(s.WorkContextID)
	}

	return nil, errors.New("no active work context")
}

func checkpointTrackingID(ctx *workcontext.WorkContext, s *Session) string {
	if s != nil && s.ID != "" {
		return s.ID
	}

	return ctx.ID
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

func collectCheckpointInput(issueID, trackingID, message string, targets CheckpointTargets, createdAt time.Time) (string, string, error) {
	if strings.TrimSpace(message) != "" {
		return strings.TrimSpace(message), defaultCommitMessage(issueID, createdAt), nil
	}

	if targets.Git {
		return openGitCheckpointEditor(issueID, trackingID, createdAt)
	}

	content, err := openEditor(checkpointTemplate(issueID, trackingID))
	if err != nil {
		return "", "", err
	}

	return trimEditorContent(content), "", nil
}

func openGitCheckpointEditor(issueID, trackingID string, createdAt time.Time) (string, string, error) {
	defaultSummary := defaultCommitMessage(issueID, createdAt)
	content, err := openEditor(gitCheckpointTemplate(issueID, trackingID, defaultSummary))
	if err != nil {
		return "", "", err
	}

	gitSummary, message := parseGitCheckpointContent(content)
	if gitSummary == "" {
		gitSummary = defaultSummary
	}

	return message, gitSummary, nil
}

func openEditor(template string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = fallbackEditor()
	}

	tmpFile, err := os.CreateTemp("", "follower-checkpoint-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

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

func checkpointTemplate(issueID, trackingID string) string {
	return fmt.Sprintf(
		"# Follower checkpoint\n# Issue: %s\n# Context: %s\n# Lines starting with # are ignored\n\n",
		issueID,
		trackingID,
	)
}

func gitCheckpointTemplate(issueID, trackingID, summary string) string {
	return fmt.Sprintf(
		"# Follower checkpoint\n# Issue: %s\n# Context: %s\n# Lines starting with # are ignored\n\n[GIT_SUMMARY]\n%s\n\n[MESSAGE]\n",
		issueID,
		trackingID,
		summary,
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

func parseGitCheckpointContent(content string) (string, string) {
	var (
		summaryLines []string
		messageLines []string
		section      string
	)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		switch trimmed {
		case "[GIT_SUMMARY]":
			section = "git_summary"
			continue
		case "[MESSAGE]":
			section = "message"
			continue
		}

		switch section {
		case "git_summary":
			summaryLines = append(summaryLines, line)
		case "message":
			messageLines = append(messageLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(summaryLines, "\n")), strings.TrimSpace(strings.Join(messageLines, "\n"))
}

func checkpointWindowStart(s *Session) time.Time {
	for i := len(s.Checkpoints) - 1; i >= 0; i-- {
		if s.Checkpoints[i].Targets.Jira {
			return s.Checkpoints[i].CreatedAt
		}
	}

	return s.StartedAt
}

func defaultCommitMessage(issueID string, createdAt time.Time) string {
	return fmt.Sprintf("chore(%s): checkpoint %s", issueID, createdAt.Format(time.RFC3339))
}

func commitGitCheckpoint(gitCheckpoint *GitCheckpoint) (string, error) {
	if gitCheckpoint == nil {
		return "", errors.New("git checkpoint is nil")
	}

	client := git.NewClient()
	return client.Commit(git.CommitInput{
		Summary: gitCheckpoint.CommitMessage,
		Body:    gitCheckpoint.CommitBody,
		Time:    gitCheckpoint.CommitTime,
	})
}

func SyncCheckpoints(ctx context.Context, jiraClient *jira.Client) error {
	s, err := loadActiveSession()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("no active session")
		}

		return err
	}

	if len(s.Checkpoints) == 0 {
		return errors.New("no checkpoints in active session")
	}

	for i := range s.Checkpoints {
		checkpoint := &s.Checkpoints[i]
		if checkpoint.Jira == nil || checkpoint.Jira.SyncedAt != nil {
			continue
		}

		comment, err := jiraClient.PostComment(ctx, s.IssueID, jiraCommentBody(s, checkpoint))
		if err != nil {
			checkpoint.SyncFailed = true
			checkpoint.Jira.LastError = err.Error()
			return saveActiveSession(s)
		}

		now := time.Now().UTC()
		checkpoint.SyncFailed = false
		checkpoint.Jira.CommentID = comment.ID
		checkpoint.Jira.SyncedAt = &now
		checkpoint.Jira.LastError = ""

		if err := saveActiveSession(s); err != nil {
			return err
		}
	}

	if hasPendingGitCheckpoints(s) {
		client := git.NewClient()
		if err := client.Push(); err != nil {
			markPendingGitCheckpointErrors(s, err.Error())
			return saveActiveSession(s)
		}

		now := time.Now().UTC()
		markPendingGitCheckpointsPushed(s, now)

		if err := saveActiveSession(s); err != nil {
			return err
		}
	}

	return nil
}

func jiraCommentBody(s *Session, checkpoint *Checkpoint) string {
	if s == nil || checkpoint == nil {
		return ""
	}

	return fmt.Sprintf(
		"Follower checkpoint recorded: %s\nFollower session: %s\n\n%s",
		checkpoint.CreatedAt.Format(time.RFC3339),
		s.ID,
		checkpoint.Message,
	)
}

func hasPendingGitCheckpoints(s *Session) bool {
	if s == nil {
		return false
	}

	for i := range s.Checkpoints {
		checkpoint := &s.Checkpoints[i]
		if checkpoint.Git != nil && checkpoint.Git.PushedAt == nil {
			return true
		}
	}

	return false
}

func markPendingGitCheckpointsPushed(s *Session, pushedAt time.Time) {
	if s == nil {
		return
	}

	for i := range s.Checkpoints {
		checkpoint := &s.Checkpoints[i]
		if checkpoint.Git == nil || checkpoint.Git.PushedAt != nil {
			continue
		}

		checkpoint.SyncFailed = false
		checkpoint.Git.PushedAt = &pushedAt
		checkpoint.Git.LastError = ""
	}
}

func markPendingGitCheckpointErrors(s *Session, lastError string) {
	if s == nil {
		return
	}

	for i := range s.Checkpoints {
		checkpoint := &s.Checkpoints[i]
		if checkpoint.Git == nil || checkpoint.Git.PushedAt != nil {
			continue
		}

		checkpoint.SyncFailed = true
		checkpoint.Git.LastError = lastError
	}
}
