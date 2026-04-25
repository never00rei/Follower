package workcontext

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	EventContextCreated       EventType = "context_created"
	EventCheckpointRecorded   EventType = "checkpoint_recorded"
	EventDistractionStarted   EventType = "distraction_started"
	EventDistractionEnded     EventType = "distraction_ended"
	EventContextPaused        EventType = "context_paused"
	EventContextResumed       EventType = "context_resumed"
	EventSyncRequested        EventType = "sync_requested"
	EventJiraCommentRequested EventType = "jira_comment_requested"
	EventJiraCommentCreated   EventType = "jira_comment_created"
	EventJiraWorklogRequested EventType = "jira_worklog_requested"
	EventJiraWorklogCreated   EventType = "jira_worklog_created"
	EventGitCommitRequested   EventType = "git_commit_requested"
	EventGitCommitCreated     EventType = "git_commit_created"
	EventSyncFailed           EventType = "sync_failed"
)

const (
	ContextSourceFollow   ContextSource = "follow"
	ContextSourceDistract ContextSource = "distract"
	ContextSourceResume   ContextSource = "resume"
)

type ContextCreatedData struct {
	ContextID       string        `json:"context_id"`
	JiraTicketID    string        `json:"jira_ticket_id"`
	ParentContextID string        `json:"parent_context_id,omitempty"`
	Source          ContextSource `json:"source"`
}

func (ContextCreatedData) eventData() {}

func NewContextCreatedEvent(id string, timestamp time.Time, data ContextCreatedData) ContextEvent {
	return ContextEvent{
		ID:        id,
		Type:      EventContextCreated,
		Version:   1,
		Timestamp: timestamp,
		Data:      data,
	}
}

func (e ContextEvent) MarshalJSON() ([]byte, error) {
	type contextEvent ContextEvent
	return json.Marshal(struct {
		contextEvent
	}{
		contextEvent: contextEvent(e),
	})
}

func (e *ContextEvent) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID        string          `json:"id"`
		Type      EventType       `json:"type"`
		Version   int             `json:"version"`
		Timestamp time.Time       `json:"timestamp"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	eventData, err := unmarshalEventData(raw.Type, raw.Data)
	if err != nil {
		return err
	}

	e.ID = raw.ID
	e.Type = raw.Type
	e.Version = raw.Version
	e.Timestamp = raw.Timestamp
	e.Data = eventData

	return nil
}

func unmarshalEventData(eventType EventType, data json.RawMessage) (EventData, error) {
	switch eventType {
	case EventContextCreated:
		var eventData ContextCreatedData
		if err := json.Unmarshal(data, &eventData); err != nil {
			return nil, err
		}

		return eventData, nil
	default:
		return nil, fmt.Errorf("unsupported event type %q", eventType)
	}
}
