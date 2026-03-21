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
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
