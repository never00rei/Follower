package jira

import "encoding/json"

// These structs build out what a Jira issue looks like.
type Issue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self"`
	Fields IssueFields `json:"fields"`
}

type IssueFields struct {
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description"`
	Status      IssueStatus     `json:"status"`
	IssueType   IssueType       `json:"issuetype"`
	Project     IssueProject    `json:"project"`
	Assignee    *IssueUser      `json:"assignee"`
	Reporter    *IssueUser      `json:"reporter"`
	Priority    *IssuePriority  `json:"priority"`
	Labels      []string        `json:"labels"`
	Created     string          `json:"created"`
	Updated     string          `json:"updated"`
}

type IssueStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

type IssueProject struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type IssueUser struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type IssuePriority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// The following structs are used in the Sync process for commentary.
type ADFDocument struct {
	Type    string    `json:"type"`
	Version int       `json:"version"`
	Content []ADFNode `json:"content"`
}

type ADFNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text,omitempty"`
	Content []ADFNode `json:"content,omitempty"`
}

type Comment struct {
	ID   string      `json:"id"`
	Body ADFDocument `json:"body"`
}

type CreateCommentRequest struct {
	Body ADFDocument `json:"body"`
}
