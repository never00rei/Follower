package session

import "time"

type Session struct {
	ID          string       `json:"id"`
	IssueID     string       `json:"issue_id"`
	StartedAt   time.Time    `json:"started_at"`
	EndedAt     *time.Time   `json:"ended_at,omitempty"`
	Checkpoints []Checkpoint `json:"checkpoints"`
}

type Checkpoint struct {
	Message    string            `json:"message"`
	CreatedAt  time.Time         `json:"created_at"`
	Targets    CheckpointTargets `json:"targets"`
	SyncFailed bool              `json:"sync_failed,omitempty"`
	Git        *GitCheckpoint    `json:"git,omitempty"`
	Jira       *JiraCheckpoint   `json:"jira,omitempty"`
}

type CheckpointTargets struct {
	Jira bool `json:"jira"`
	Git  bool `json:"git"`
}

type GitCheckpoint struct {
	CommitMessage string    `json:"commit_message"`
	CommitTime    time.Time `json:"commit_time"`
	CommitBody    string    `json:"commit_body,omitempty"`
	CommitSHA     string    `json:"commit_sha,omitempty"`
	PushedAt      *time.Time `json:"pushed_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
}

type JiraCheckpoint struct {
	WindowStartedAt time.Time `json:"window_started_at"`
	WindowEndedAt   time.Time `json:"window_ended_at"`
	CommentID       string    `json:"comment_id,omitempty"`
	SyncedAt        *time.Time `json:"synced_at,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
}
