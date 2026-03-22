package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type CommitInput struct {
	Summary string
	Body    string
	Time    time.Time
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Commit(input CommitInput) (string, error) {
	if err := ensureStagedChanges(); err != nil {
		return "", err
	}

	args := []string{"commit", "-m", input.Summary}
	if strings.TrimSpace(input.Body) != "" {
		args = append(args, "-m", input.Body)
	}

	commitTime := input.Time.Format(time.RFC3339)
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE="+commitTime,
		"GIT_COMMITTER_DATE="+commitTime,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git commit failed: %w", err)
	}

	shaCmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := shaCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func ensureStagedChanges() error {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	if err := cmd.Run(); err == nil {
		return errors.New("no staged changes to commit")
	} else {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil
		}

		return fmt.Errorf("could not inspect staged changes: %w", err)
	}
}
