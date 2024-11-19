package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	err := run()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Check if we're in a git repository
	if err := checkGitRepo(); err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// Check if this is the first commit
	isFirstCommit, err := isRepositoryEmpty()
	if err != nil {
		return fmt.Errorf("failed to check if repository is empty: %w", err)
	}

	if isFirstCommit {
		return createFirstCommit()
	}

	// Try to fetch from remote if there's a remote origin

	if err := checkOriginRemote(); err != nil {
		return fmt.Errorf("no remote origin found: %w", err)
	}

	if err := fetchFromRemote(); err != nil {
		fmt.Println("No remote found or fetch failed, continuing with local operations")
	} else {
		// If fetch succeeded, try to pull
		if err := pullChanges(); err != nil {
			return fmt.Errorf("failed to pull changes: %w", err)
		}
	}

	// Get next commit number
	nextNumber, err := determineNextCommitNumber()
	if err != nil {
		return fmt.Errorf("failed to determine next commit number: %w", err)
	}

	// Create the commit
	if err := createCommit(nextNumber); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	fmt.Printf("Successfully created commit %d\n", nextNumber)
	return nil
}

func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run()
}

func isRepositoryEmpty() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	if err := cmd.Run(); err != nil {
		// If HEAD doesn't exist, this is an empty repository
		return true, nil
	}
	return false, nil
}

func createFirstCommit() error {
	// Stage all files
	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Create first commit
	cmd = exec.Command("git", "commit", "-m", "1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create first commit: %s", out)
	}
	return nil
}

func checkOriginRemote() error {
	cmd := exec.Command("git", "remote")
	remotes := cmd.Run()
	return remotes
}

func fetchFromRemote() error {
	cmd := exec.Command("git", "fetch")
	return cmd.Run()
}

func pullChanges() error {
	cmd := exec.Command("git", "pull")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull failed: %s", out)
	}
	return nil
}

func determineNextCommitNumber() (int, error) {
	// Try to get the last commit message
	lastMsg, err := getLastCommitMessage()
	if err != nil {
		return 0, fmt.Errorf("failed to get last commit message: %w", err)
	}

	// Try to parse it as a number
	if num, err := strconv.Atoi(strings.TrimSpace(lastMsg)); err == nil {
		return num + 1, nil
	}

	// If parsing failed, count all non-merge commits
	return countNonMergeCommits()
}

func getLastCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func countNonMergeCommits() (int, error) {
	// Get all commits but exclude merges
	cmd := exec.Command("git", "log", "--no-merges", "--oneline")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get git log: %w", err)
	}

	// Count lines, each line is one commit
	commits := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := len(commits)
	if commits[0] == "" {
		count = 0
	}
	return count + 1, nil
}

func createCommit(number int) error {
	// Stage all changes
	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create commit
	cmd = exec.Command("git", "commit", "-m", strconv.Itoa(number))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create commit: %s", out)
	}
	return nil
}
