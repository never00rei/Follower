package workcontext

import "time"

type WorkContext struct {
	ID              string         `json:"id"`
	JiraTicketID    string         `json:"jira_ticket_id"`
	ParentContextID string         `json:"parent_context_id,omitempty"`
	Events          []ContextEvent `json:"events"`
}

type ContextEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Data      EventData `json:"data"`
}

type EventType string

type EventData interface {
	eventData()
}

type ContextSource string

type CurrentContextRef struct {
	ContextID  string        `json:"context_id"`
	UpdatedAt  time.Time     `json:"updated_at"`
	LastSource ContextSource `json:"last_source"`
}

type ContextHistory struct {
	Contexts []ContextHistoryEntry `json:"contexts"`
}

type ContextHistoryEntry struct {
	ContextID       string    `json:"context_id"`
	JiraTicketID    string    `json:"jira_ticket_id"`
	ParentContextID string    `json:"parent_context_id,omitempty"`
	LastActiveAt    time.Time `json:"last_active_at"`
}
