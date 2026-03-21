package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/never00rei/Follower/internal/config"
	"github.com/never00rei/Follower/internal/jira"
	"github.com/never00rei/Follower/internal/session"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "follower:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "init":
		return initConfig(os.Stdin, os.Stdout)
	case "follow":
		if len(args) != 2 {
			return fmt.Errorf("usage: follower follow ISSUE-ID")
		}

		return followIssue(args[1], os.Stdout)
	case "issue":
		if len(args) != 2 {
			return fmt.Errorf("usage: follower issue ISSUE-ID")
		}

		return showIssue(args[1], os.Stdout)
	case "checkpoint":
		message := strings.TrimSpace(strings.Join(args[1:], " "))
		return session.AddCheckpoint(message)
	case "status":
		return session.Status(os.Stdout)
	case "done":
		return session.Done()
	case "help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage() {
	fmt.Println("Follower")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  follower init")
	fmt.Println("  follower follow ISSUE-ID")
	fmt.Println("  follower issue ISSUE-ID")
	fmt.Println("  follower checkpoint [MESSAGE]")
	fmt.Println("  follower status")
	fmt.Println("  follower done")
}

func showIssue(issueID string, stdout *os.File) error {
	conf, err := config.LoadConfiguration()
	if err != nil {
		return err
	}

	client, err := jira.NewClient(conf.Atlassian)
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(context.Background(), issueID)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(issue)
}

func followIssue(issueID string, stdout *os.File) error {
	conf, err := config.LoadConfiguration()
	if err != nil {
		return err
	}

	client, err := jira.NewClient(conf.Atlassian)
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(context.Background(), issueID)
	if err != nil {
		return fmt.Errorf("could not load Jira issue %s: %w", issueID, err)
	}

	if _, err := fmt.Fprintf(stdout, "%s [%s] %s\n", issue.Key, issue.Fields.Status.Name, issue.Fields.Summary); err != nil {
		return err
	}

	return session.Follow(issue.Key, stdout)
}

func initConfig(stdin *os.File, stdout *os.File) error {
	conf, err := config.LoadConfiguration()
	if err != nil {
		return err
	}

	reader := bufio.NewReader(stdin)

	fmt.Fprintln(stdout, "Follower Atlassian setup")
	fmt.Fprintln(stdout, "")

	siteURL, err := prompt(reader, stdout, "Atlassian site URL", conf.Atlassian.AtlassianSiteURL)
	if err != nil {
		return err
	}

	userEmail, err := prompt(reader, stdout, "Atlassian user email", conf.Atlassian.UserEmail)
	if err != nil {
		return err
	}

	apiKey, err := prompt(reader, stdout, "Atlassian API key", conf.Atlassian.AtlassianAPIKey)
	if err != nil {
		return err
	}

	cloudID, err := config.ResolveCloudID(siteURL)
	if err != nil {
		return err
	}

	conf.Atlassian.AtlassianSiteURL = siteURL
	conf.Atlassian.AtlassianCloudID = cloudID
	conf.Atlassian.UserEmail = userEmail
	conf.Atlassian.AtlassianAPIKey = apiKey

	if err := config.SaveConfiguration(conf); err != nil {
		return err
	}

	path, err := config.FilePath()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "\nResolved Atlassian cloud ID: %s\nSaved configuration to %s\n", cloudID, path)
	return err
}

func prompt(reader *bufio.Reader, stdout *os.File, label, current string) (string, error) {
	if current != "" {
		if _, err := fmt.Fprintf(stdout, "%s [%s]: ", label, current); err != nil {
			return "", err
		}
	} else {
		if _, err := fmt.Fprintf(stdout, "%s: ", label); err != nil {
			return "", err
		}
	}

	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return current, nil
	}

	return value, nil
}
