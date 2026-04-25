package workcontext

import (
	"errors"
	"strings"
	"time"
)

func Create(id, eventID, jiraTicketID, parentContextID string, source ContextSource, timestamp time.Time) (*WorkContext, error) {
	id = strings.TrimSpace(id)
	eventID = strings.TrimSpace(eventID)
	jiraTicketID = strings.TrimSpace(jiraTicketID)
	parentContextID = strings.TrimSpace(parentContextID)

	if id == "" {
		return nil, errors.New("work context id is required")
	}

	if eventID == "" {
		return nil, errors.New("context created event id is required")
	}

	if jiraTicketID == "" {
		return nil, errors.New("jira ticket id is required")
	}

	if source == "" {
		return nil, errors.New("context source is required")
	}

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
