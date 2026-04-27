package workcontext

import (
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

func NewContextID(timestamp time.Time) string {
	return "ctx_" + ulid.MustNew(ulid.Timestamp(timestamp), ulid.DefaultEntropy()).String()
}

func NewEventID(timestamp time.Time) string {
	return "evt_" + ulid.MustNew(ulid.Timestamp(timestamp), ulid.DefaultEntropy()).String()
}

func Create(jiraTicketID, parentContextID string, source ContextSource, timestamp time.Time) (*WorkContext, error) {
	jiraTicketID = strings.TrimSpace(jiraTicketID)
	parentContextID = strings.TrimSpace(parentContextID)

	if jiraTicketID == "" {
		return nil, errors.New("jira ticket id is required")
	}

	if source == "" {
		return nil, errors.New("context source is required")
	}

	id := NewContextID(timestamp)
	eventID := NewEventID(timestamp)

	ctx := &WorkContext{
		ID:              id,
		JiraTicketID:    jiraTicketID,
		ParentContextID: parentContextID,
		Events: []ContextEvent{
			NewContextCreatedEvent(eventID, timestamp, ContextCreatedData{
				ContextID:       id,
				JiraTicketID:    jiraTicketID,
				ParentContextID: parentContextID,
				Source:          source,
			}),
		},
	}

	if err := Save(ctx); err != nil {
		return nil, err
	}

	if err := SaveCurrent(CurrentContextRef{
		ContextID:  id,
		UpdatedAt:  timestamp,
		LastSource: source,
	}); err != nil {
		return nil, err
	}

	if err := UpsertHistoryEntry(ContextHistoryEntry{
		ContextID:       id,
		JiraTicketID:    jiraTicketID,
		ParentContextID: parentContextID,
		LastActiveAt:    timestamp,
	}); err != nil {
		return nil, err
	}

	return ctx, nil
}

func AppendEvent(contextID string, event ContextEvent) (*WorkContext, error) {
	contextID = strings.TrimSpace(contextID)
	if contextID == "" {
		return nil, errors.New("work context id is required")
	}

	if err := validateEvent(event); err != nil {
		return nil, err
	}

	ctx, err := Load(contextID)
	if err != nil {
		return nil, err
	}

	ctx.Events = append(ctx.Events, event)

	if err := Save(ctx); err != nil {
		return nil, err
	}

	current := CurrentContextRef{
		ContextID: contextID,
		UpdatedAt: event.Timestamp,
	}

	if existing, err := LoadCurrent(); err == nil {
		current.LastSource = existing.LastSource
	}

	if err := SaveCurrent(current); err != nil {
		return nil, err
	}

	if err := UpsertHistoryEntry(ContextHistoryEntry{
		ContextID:       ctx.ID,
		JiraTicketID:    ctx.JiraTicketID,
		ParentContextID: ctx.ParentContextID,
		LastActiveAt:    event.Timestamp,
	}); err != nil {
		return nil, err
	}

	return ctx, nil
}

func RecordCheckpoint(contextID, message string, targets CheckpointTargets, timestamp time.Time) (*WorkContext, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, errors.New("checkpoint message is required")
	}

	return AppendEvent(contextID, NewCheckpointRecordedEvent(
		NewEventID(timestamp),
		timestamp,
		CheckpointRecordedData{
			Message: message,
			Targets: targets,
		},
	))
}

func validateEvent(event ContextEvent) error {
	if strings.TrimSpace(event.ID) == "" {
		return errors.New("event id is required")
	}

	if event.Type == "" {
		return errors.New("event type is required")
	}

	if event.Version == 0 {
		return errors.New("event version is required")
	}

	if event.Timestamp.IsZero() {
		return errors.New("event timestamp is required")
	}

	if event.Data == nil {
		return errors.New("event data is required")
	}

	return nil
}
